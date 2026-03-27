from moto import mock_aws
import boto3
from novabackup.aws_real import AWSCloudProvider


@mock_aws
def test_aws_cloud_provider_list_and_backup():
    # Create a mock EC2 instance with a Name tag
    ec2 = boto3.client("ec2", region_name="us-east-1")
    ec2.run_instances(
        ImageId="ami-12345678",
        MinCount=1,
        MaxCount=1,
        TagSpecifications=[
            {
                "ResourceType": "instance",
                "Tags": [{"Key": "Name", "Value": "test-aws-vm"}],
            }
        ],
    )

    provider = AWSCloudProvider(region_name="us-east-1")
    vms = provider.list_vms()
    assert isinstance(vms, list)
    # Attempt a cloud backup (will rely on moto mocking of AWS APIs)
    if vms:
        vm = vms[0]
        resp = provider.backup_to_cloud(
            vm_id=vm.get("name") or vm.get("id"),
            cloud_provider="AWS",
            region="us-east-1",
            dest="s3://bucket",
            backup_type="full",
            snapshot_name="test-snap",
        )
        assert isinstance(resp, dict) or resp is None
