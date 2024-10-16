package publisher_agent

type PublisherMetadata struct {
	IsAws                 bool   `json:"is_aws"`
	AwsInstanceType       string `json:"aws_instance_type"`
	AwsAvailabilityZoneId string `json:"aws_availability_zone_id"`
	PublisherAgentVersion string `json:"publisher_agent_version"`
	Architecture          string `json:"architecture"`
	PublicIp              string `json:"public_ip"`
}

const awsMetadataEndpoint = "http://169.254.169.254/latest/meta-data"

func GetPublisherMetadata() PublisherMetadata {
	instanceTypeResult, err := RestQuery("GET", awsMetadataEndpoint+"/instance-type", nil, nil, nil)
	if err != nil {
		return PublisherMetadata{
			IsAws: false,
		}
	}
	instanceType := string(instanceTypeResult)

	azResult, err := RestQuery("GET", awsMetadataEndpoint+"/availability-zone-id", nil, nil, nil)
	if err != nil {
		return PublisherMetadata{
			IsAws: false,
		}
	}
	availabilityZone := string(azResult)

	return PublisherMetadata{
		IsAws:                 true,
		AwsAvailabilityZoneId: availabilityZone,
		AwsInstanceType:       instanceType,
	}
}
