"""Cloud provider integration module (real+mock).
This folder contains implementations for cloud VM backups/restores to cloud providers such as AWS, Azure, and Google Cloud.
Current MVP includes a Mock provider for testing/CI readiness and real providers when dependencies are installed.
"""

from .aws import AWSCloudProvider  # type: ignore
from .azure import AzureCloudProvider  # type: ignore
from .gcp import GCPCloudProvider  # type: ignore
from .mock import MockCloudProvider  # type: ignore

__all__ = [
    "AWSCloudProvider",
    "AzureCloudProvider",
    "GCPCloudProvider",
    "MockCloudProvider",
]
