"""
Hyper-V Connection Module
Uses PowerShell to interact with Hyper-V
"""
import subprocess
import json
import logging
from typing import List, Dict, Optional, Any
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass
class HyperVConfig:
    """Hyper-V connection configuration"""
    host: str = "localhost"  # Local or remote host
    auth_method: str = "local"  # "local" or "remote"
    username: Optional[str] = None
    password: Optional[str] = None
    timeout: int = 60


class HyperVConnection:
    """Hyper-V connection manager using PowerShell"""
    
    def __init__(self, config: HyperVConfig):
        self.config = config
        self.connected = False
    
    def _run_powershell(self, script: str) -> Dict[str, Any]:
        """Execute PowerShell script and return JSON output"""
        try:
            # Wrap script to output JSON
            full_script = f"""
            $ErrorActionPreference = "Stop"
            try {{
                {script}
            }} catch {{
                $result = @{{
                    Success = $false
                    Error = $_.Exception.Message
                }}
                $result | ConvertTo-Json -Depth 5
            }}
            """
            
            # Execute PowerShell
            process = subprocess.Popen(
                ["powershell", "-Command", full_script],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                encoding="utf-8",
                errors="replace"
            )
            
            stdout, stderr = process.communicate(timeout=self.config.timeout)
            
            if process.returncode != 0:
                logger.error(f"PowerShell error: {stderr}")
                return {"Success": False, "Error": stderr}
            
            # Parse JSON output
            try:
                result = json.loads(stdout.strip())
                return result
            except json.JSONDecodeError:
                logger.error(f"Failed to parse JSON output: {stdout}")
                return {"Success": False, "Error": "Failed to parse output", "RawOutput": stdout}
                
        except subprocess.TimeoutExpired:
            logger.error("PowerShell command timed out")
            return {"Success": False, "Error": "Command timed out"}
        except Exception as e:
            logger.error(f"Error running PowerShell: {str(e)}")
            return {"Success": False, "Error": str(e)}
    
    def test_connection(self) -> Dict[str, Any]:
        """Test Hyper-V connection"""
        result = {
            "success": False,
            "host": self.config.host,
            "message": "",
            "hyper_v_enabled": False,
            "version": None,
        }
        
        try:
            # Check if Hyper-V role is installed
            script = """
            $hyperV = Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V -Online -ErrorAction SilentlyContinue
            $result = @{
                Success = $true
                HyperVEnabled = $hyperV.State -eq "Enabled"
                State = $hyperV.State
            }
            $result | ConvertTo-Json -Depth 3
            """
            
            response = self._run_powershell(script)
            
            if response.get("Success"):
                result["success"] = True
                result["hyper_v_enabled"] = response.get("HyperVEnabled", False)
                result["state"] = response.get("State", "Unknown")
                result["message"] = "Hyper-V is available" if result["hyper_v_enabled"] else "Hyper-V is not enabled"
            else:
                result["message"] = response.get("Error", "Unknown error")
                
        except Exception as e:
            result["message"] = str(e)
            logger.error(f"Connection test failed: {str(e)}")
        
        return result
    
    def list_vms(self) -> List[Dict[str, Any]]:
        """List all virtual machines on Hyper-V host"""
        vms = []
        
        try:
            # PowerShell script to list VMs
            script = """
            $vms = Get-VM -ErrorAction SilentlyContinue | ForEach-Object {
                @{
                    Success = $true
                    VMs = @(
                        $_ | Select-Object Name, Id, State, Generation, 
                            @{Name="MemoryStartup";Expression={$_.MemoryStartup}},
                            @{Name="MemoryMinimum";Expression={$_.MemoryMinimum}},
                            @{Name="MemoryMaximum";Expression={$_.MemoryMaximum}},
                            @{Name="ProcessorCount";Expression={$_.ProcessorCount}},
                            @{Name="IntegrationServicesState";Expression={$_.IntegrationServicesState}},
                            @{Name="Uptime";Expression={$_.Uptime}},
                            @{Name="CheckpointType";Expression={$_.CheckpointType}}
                    )
                }
            }
            
            if ($vms) {
                $vms | ConvertTo-Json -Depth 5
            } else {
                @{Success = $true; VMs = @()} | ConvertTo-Json
            }
            """
            
            response = self._run_powershell(script)
            
            if response.get("Success") and response.get("VMs"):
                for vm in response["VMs"]:
                    vm_info = {
                        "id": str(vm.get("Id", "")),
                        "name": vm.get("Name", ""),
                        "state": vm.get("State", "Unknown"),
                        "generation": vm.get("Generation", 0),
                        "memory_startup_mb": int(vm.get("MemoryStartup", 0)) // 1024 // 1024 if vm.get("MemoryStartup") else 0,
                        "memory_minimum_mb": int(vm.get("MemoryMinimum", 0)) // 1024 // 1024 if vm.get("MemoryMinimum") else 0,
                        "memory_maximum_mb": int(vm.get("MemoryMaximum", 0)) // 1024 // 1024 if vm.get("MemoryMaximum") else 0,
                        "processor_count": vm.get("ProcessorCount", 0),
                        "integration_services": vm.get("IntegrationServicesState", "Unknown"),
                        "uptime_ticks": vm.get("Uptime", 0),
                        "checkpoint_type": vm.get("CheckpointType", "Disabled"),
                    }
                    
                    # Calculate uptime in human-readable format
                    uptime_ticks = vm_info["uptime_ticks"]
                    if uptime_ticks:
                        uptime_seconds = uptime_ticks / 10000000  # Convert from ticks to seconds
                        uptime_hours = int(uptime_seconds / 3600)
                        vm_info["uptime_hours"] = uptime_hours
                    else:
                        vm_info["uptime_hours"] = 0
                    
                    vms.append(vm_info)
                
                logger.info(f"Found {len(vms)} VMs on Hyper-V host")
            
        except Exception as e:
            logger.error(f"Error listing VMs: {str(e)}")
        
        return vms
    
    def get_vm_info(self, vm_name: str) -> Optional[Dict[str, Any]]:
        """Get detailed information about a specific VM"""
        try:
            script = f"""
            $vm = Get-VM -Name "{vm_name}" -ErrorAction SilentlyContinue
            if ($vm) {{
                $disks = Get-VMHardDiskDrive -VMName "{vm_name}" -ErrorAction SilentlyContinue | ForEach-Object {{
                    @{{
                        ControllerType = $_.ControllerType
                        ControllerNumber = $_.ControllerNumber
                        ControllerLocation = $_.ControllerLocation
                        Path = $_.Path
                        Size = $_.Size
                    }}
                }}
                
                $network = Get-VMNetworkAdapter -VMName "{vm_name}" -ErrorAction SilentlyContinue | ForEach-Object {{
                    @{{
                        Name = $_.Name
                        SwitchName = $_.SwitchName
                        MacAddress = $_.MacAddress
                        IpAddresses = $_.IPAddresses
                    }}
                }}
                
                $result = @{{
                    Success = $true
                    VM = $vm | Select-Object Name, Id, State, Generation, 
                        @{{Name="MemoryStartup";Expression={{$_.MemoryStartup}}}},
                        @{{Name="MemoryMinimum";Expression={{$_.MemoryMinimum}}}},
                        @{{Name="MemoryMaximum";Expression={{$_.MemoryMaximum}}}},
                        @{{Name="ProcessorCount";Expression={{$_.ProcessorCount}}}},
                        @{{Name="IntegrationServicesState";Expression={{$_.IntegrationServicesState}}}},
                        @{{Name="Uptime";Expression={{$_.Uptime}}}},
                        @{{Name="CheckpointType";Expression={{$_.CheckpointType}}}},
                        @{{Name="CPUUsage";Expression={{$_.CPUUsage}}}},
                        @{{Name="MemoryAssigned";Expression={{$_.MemoryAssigned}}}}
                    Disks = $disks
                    Network = $network
                }}
                $result | ConvertTo-Json -Depth 5
            }} else {{
                @{{Success = $false; Error = "VM not found"}} | ConvertTo-Json
            }}
            """
            
            response = self._run_powershell(script)
            
            if response.get("Success") and response.get("VM"):
                vm = response["VM"]
                return {
                    "name": vm.get("Name", ""),
                    "state": vm.get("State", "Unknown"),
                    "generation": vm.get("Generation", 0),
                    "memory_startup_mb": int(vm.get("MemoryStartup", 0)) // 1024 // 1024,
                    "processor_count": vm.get("ProcessorCount", 0),
                    "cpu_usage": int(vm.get("CPUUsage", 0)),
                    "memory_assigned_mb": int(vm.get("MemoryAssigned", 0)) // 1024 // 1024,
                    "disks": response.get("Disks", []),
                    "network": response.get("Network", []),
                }
            
        except Exception as e:
            logger.error(f"Error getting VM info: {str(e)}")
        
        return None
    
    def create_checkpoint(self, vm_name: str, checkpoint_name: str) -> Dict[str, Any]:
        """Create a checkpoint for a VM"""
        result = {
            "success": False,
            "vm_name": vm_name,
            "checkpoint_name": checkpoint_name,
            "message": "",
        }
        
        try:
            script = f"""
            $checkpoint = Checkpoint-VM -Name "{vm_name}" -SnapshotName "{checkpoint_name}" -ErrorAction SilentlyContinue
            if ($checkpoint) {{
                @{{
                    Success = $true
                    Message = "Checkpoint created successfully"
                    Name = $checkpoint.Name
                    Created = $checkpoint.CreationTime
                }} | ConvertTo-Json
            }} else {{
                @{{Success = $false; Error = "Failed to create checkpoint"}} | ConvertTo-Json
            }}
            """
            
            response = self._run_powershell(script)
            result["success"] = response.get("Success", False)
            result["message"] = response.get("Message", response.get("Error", "Unknown error"))
            
        except Exception as e:
            result["message"] = str(e)
            logger.error(f"Error creating checkpoint: {str(e)}")
        
        return result


def connect_to_hyperv(
    host: str = "localhost",
    auth_method: str = "local",
    username: Optional[str] = None,
    password: Optional[str] = None,
) -> HyperVConnection:
    """
    Create Hyper-V connection
    
    Args:
        host: Hyper-V host name or IP (default: localhost)
        auth_method: "local" or "remote"
        username: Username for remote connection
        password: Password for remote connection
    
    Returns:
        HyperVConnection instance
    """
    config = HyperVConfig(
        host=host,
        auth_method=auth_method,
        username=username,
        password=password,
    )
    return HyperVConnection(config)
