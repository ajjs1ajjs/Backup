import json
import platform
import subprocess
from typing import List, Dict


class HyperVProvider:
    def list_vms(self) -> List[Dict[str, str]]:
        # Only on Windows Hyper-V host. On non-Windows, return empty list.
        if platform.system().lower() != "windows":
            return []
        cmd = [
            "powershell",
            "-NoProfile",
            "-ExecutionPolicy",
            "Bypass",
            "Get-VM | Select-Object Name,Id,State | ConvertTo-Json",
        ]
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        except Exception:
            return []
        out = result.stdout.strip()
        if not out:
            return []
        try:
            data = json.loads(out)
        except json.JSONDecodeError:
            return []
        if isinstance(data, dict):
            data = [data]
        vms: List[Dict[str, str]] = []
        for vm in data:
            name = vm.get("Name") or vm.get("name") or vm.get("VMName") or ""
            vid = vm.get("Id") or vm.get("ID") or vm.get("UUID") or vm.get("VMId") or ""
            state = vm.get("State") or vm.get("Status") or ""
            state_l = str(state).lower()
            if state_l in {"running", "started"}:
                status = "running"
            elif state_l in {"stopped", "stopping", "off"}:
                status = "stopped"
            else:
                status = "unknown"
            vms.append(
                {"id": str(vid), "name": str(name), "type": "HyperV", "status": status}
            )
        return vms
