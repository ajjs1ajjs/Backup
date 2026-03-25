from typing import List, Dict

try:
    import libvirt  # type: ignore
except Exception:  # pragma: no cover
    libvirt = None  # type: ignore

import platform


class LibvirtProvider:
    def list_vms(self) -> List[Dict[str, str]]:
        if libvirt is None:
            return []
        try:
            conn = libvirt.open(None)
        except libvirt.libvirtError:
            return []
        vms: List[Dict[str, str]] = []
        try:
            domains = conn.listAllDomains(0)
            for d in domains:
                try:
                    name = d.name()
                except Exception:
                    name = "unknown"
                try:
                    vid = d.UUIDString()
                except Exception:
                    vid = "0"
                state, _ = d.state()
                status = "running" if state == libvirt.VIR_DOMAIN_RUNNING else "stopped"
                vms.append(
                    {"id": str(vid), "name": str(name), "type": "KVM", "status": status}
                )
        finally:
            try:
                conn.close()
            except Exception:
                pass
        return vms
