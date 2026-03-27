"""
SQL Server Connection Module
Supports Windows Authentication and SQL Authentication
"""
import pyodbc
from typing import List, Dict, Optional, Any
from dataclasses import dataclass
import logging

logger = logging.getLogger(__name__)


@dataclass
class SQLServerConfig:
    """SQL Server connection configuration"""
    server: str
    port: int = 1433
    auth_method: str = "windows"  # "windows" or "sql"
    username: Optional[str] = None
    password: Optional[str] = None
    database: Optional[str] = None
    timeout: int = 30
    encrypt: bool = False
    trust_server_certificate: bool = True


class SQLServerConnection:
    """SQL Server connection manager"""
    
    def __init__(self, config: SQLServerConfig):
        self.config = config
        self.connection: Optional[pyodbc.Connection] = None
    
    def _build_connection_string(self) -> str:
        """Build ODBC connection string"""
        # Build server string
        if self.config.port != 1433:
            server_string = f"{self.config.server},{self.config.port}"
        else:
            server_string = self.config.server
        
        # Build connection parameters
        params = [
            f"DRIVER={{ODBC Driver 17 for SQL Server}}",
            f"SERVER={server_string}",
        ]
        
        # Authentication
        if self.config.auth_method == "windows":
            params.append("Trusted_Connection=yes")
        else:
            params.append(f"UID={self.config.username}")
            params.append(f"PWD={self.config.password}")
        
        # Database
        if self.config.database:
            params.append(f"DATABASE={self.config.database}")
        
        # Connection options
        params.append(f"CONNECT TIMEOUT={self.config.timeout}")
        params.append(f"LOGIN TIMEOUT={self.config.timeout}")
        
        if self.config.encrypt:
            params.append("Encrypt=yes")
            if self.config.trust_server_certificate:
                params.append("TrustServerCertificate=yes")
        else:
            params.append("Encrypt=no")
        
        return ";".join(params)
    
    def connect(self) -> bool:
        """Establish connection to SQL Server"""
        try:
            conn_string = self._build_connection_string()
            logger.debug(f"Connecting to SQL Server: {self.config.server}")
            self.connection = pyodbc.connect(conn_string)
            logger.info(f"Successfully connected to SQL Server: {self.config.server}")
            return True
        except pyodbc.Error as e:
            logger.error(f"SQL Server connection error: {str(e)}")
            self.connection = None
            return False
        except Exception as e:
            logger.error(f"Unexpected error connecting to SQL Server: {str(e)}")
            self.connection = None
            return False
    
    def disconnect(self):
        """Close connection"""
        if self.connection:
            self.connection.close()
            self.connection = None
            logger.info("Disconnected from SQL Server")
    
    def test_connection(self) -> Dict[str, Any]:
        """Test connection and return status"""
        result = {
            "success": False,
            "server": self.config.server,
            "message": "",
            "version": None,
        }
        
        try:
            if not self.connect():
                result["message"] = "Failed to connect"
                return result
            
            # Get SQL Server version
            cursor = self.connection.cursor()
            cursor.execute("SELECT @@VERSION")
            row = cursor.fetchone()
            if row:
                result["version"] = row[0]
            
            cursor.close()
            result["success"] = True
            result["message"] = "Connection successful"
            
        except Exception as e:
            result["message"] = str(e)
            logger.error(f"Connection test failed: {str(e)}")
        
        finally:
            self.disconnect()
        
        return result
    
    def list_databases(self) -> List[Dict[str, Any]]:
        """List all databases on the server"""
        databases = []
        
        try:
            if not self.connect():
                return databases
            
            cursor = self.connection.cursor()
            
            # Query to list databases (excluding system databases)
            query = """
                SELECT 
                    name,
                    database_id,
                    create_date,
                    state_desc,
                    user_access_desc,
                    recovery_model_desc,
                    CAST(SUM(size) * 8.0 / 1024 AS DECIMAL(10,2)) as size_mb
                FROM sys.databases d
                LEFT JOIN sys.master_files mf ON d.database_id = mf.database_id
                WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb', 'distribution')
                GROUP BY name, database_id, create_date, state_desc, user_access_desc, recovery_model_desc
                ORDER BY name
            """
            
            cursor.execute(query)
            columns = [column[0] for column in cursor.description]
            
            for row in cursor.fetchall():
                db_info = dict(zip(columns, row))
                databases.append({
                    "id": str(db_info.get("database_id", "")),
                    "name": db_info.get("name", ""),
                    "size_mb": float(db_info.get("size_mb", 0)) if db_info.get("size_mb") else 0,
                    "size_gb": round(float(db_info.get("size_mb", 0)) / 1024, 2) if db_info.get("size_mb") else 0,
                    "state": db_info.get("state_desc", "UNKNOWN"),
                    "recovery_model": db_info.get("recovery_model_desc", "UNKNOWN"),
                    "created": db_info.get("create_date", "").isoformat() if db_info.get("create_date") else None,
                })
            
            cursor.close()
            logger.info(f"Found {len(databases)} databases on {self.config.server}")
            
        except Exception as e:
            logger.error(f"Error listing databases: {str(e)}")
        
        finally:
            self.disconnect()
        
        return databases
    
    def get_database_info(self, database_name: str) -> Optional[Dict[str, Any]]:
        """Get detailed information about a specific database"""
        try:
            if not self.connect():
                return None
            
            cursor = self.connection.cursor()
            
            # Use the target database
            cursor.execute(f"USE [{database_name}]")
            
            # Get database info
            query = """
                SELECT 
                    DB_NAME() as name,
                    DATABASEPROPERTYEX(DB_NAME(), 'Status') as status,
                    DATABASEPROPERTYEX(DB_NAME(), 'UserAccess') as user_access,
                    DATABASEPROPERTYEX(DB_NAME(), 'Recovery') as recovery_model,
                    DATABASEPROPERTYEX(DB_NAME(), 'Collation') as collation,
                    (SELECT SUM(size) * 8.0 / 1024 FROM sys.database_files) as size_mb
            """
            
            cursor.execute(query)
            row = cursor.fetchone()
            
            if row:
                return {
                    "name": row.name,
                    "status": row.status,
                    "user_access": row.user_access,
                    "recovery_model": row.recovery_model,
                    "collation": row.collation,
                    "size_mb": round(float(row.size_mb), 2) if row.size_mb else 0,
                    "size_gb": round(float(row.size_mb) / 1024, 2) if row.size_mb else 0,
                }
            
            cursor.close()
            
        except Exception as e:
            logger.error(f"Error getting database info: {str(e)}")
        
        finally:
            self.disconnect()
        
        return None


def connect_to_sql_server(
    server: str,
    port: int = 1433,
    auth_method: str = "windows",
    username: Optional[str] = None,
    password: Optional[str] = None,
    database: Optional[str] = None,
) -> SQLServerConnection:
    """
    Create SQL Server connection
    
    Args:
        server: Server name or IP address
        port: SQL Server port (default: 1433)
        auth_method: "windows" or "sql"
        username: SQL username (for SQL auth)
        password: SQL password (for SQL auth)
        database: Specific database to connect to
    
    Returns:
        SQLServerConnection instance
    """
    config = SQLServerConfig(
        server=server,
        port=port,
        auth_method=auth_method,
        username=username,
        password=password,
        database=database,
    )
    return SQLServerConnection(config)
