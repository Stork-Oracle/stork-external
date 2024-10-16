package publisher_agent

import (
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/Stork-Oracle/stork_external/lib/signer"
	"github.com/rs/zerolog"
)

const publicIpUrl = "https://api.ipify.org"
const awsMetadataUrl = "http://169.254.169.254/latest/meta-data"
const publisherMetadataReportUrl = ""
const versionFile = "version.txt"

type PublisherMetadata struct {
	PublisherAgentVersion string      `json:"publisher_agent_version"`
	Architecture          string      `json:"architecture"`
	PublicIp              string      `json:"public_ip"`
	AwsMetadata           AwsMetadata `json:"aws_metadata"`
}

type AwsMetadata struct {
	IsAws              bool   `json:"is_aws"`
	InstanceType       string `json:"instance_type"`
	AvailabilityZoneId string `json:"availability_zone_id"`
}

type PublisherMetadataReporter struct {
	publicKey     signer.PublisherKey
	signatureType signer.SignatureType
	reportPeriod  time.Duration
	logger        zerolog.Logger
}

func NewPublisherMetadataReporter(
	publicKey signer.PublisherKey,
	signatureType signer.SignatureType,
	reportPeriod time.Duration,
) *PublisherMetadataReporter {
	return &PublisherMetadataReporter{
		publicKey:     publicKey,
		signatureType: signatureType,
		reportPeriod:  reportPeriod,
	}
}

func (p *PublisherMetadataReporter) Run() {
	err := p.report()
	if err != nil {
		p.logger.Warn().Msgf("Error reporting publisher metadata: %v", err)
	}
	for range time.Tick(p.reportPeriod) {
		err = p.report()
		if err != nil {
			p.logger.Warn().Msgf("Error reporting publisher metadata: %v", err)
		}
	}
}

func (p *PublisherMetadataReporter) report() error {
	metadata := p.getMetadata()
	//instanceTypeResult, err := RestQuery("GET", publisherMetadataReportUrl+"/publisher_metadata", nil, metadata, nil)
	p.logger.Debug().Msgf("Reported publisher metadata: %v", metadata)
	return nil
}

func (p *PublisherMetadataReporter) getMetadata() PublisherMetadata {
	awsMetadata := getAwsMetadata()
	architecture := runtime.GOARCH
	publicIp := getPublicIp()
	version := getPublisherAgentVersion()

	return PublisherMetadata{
		PublisherAgentVersion: version,
		Architecture:          architecture,
		PublicIp:              publicIp,
		AwsMetadata:           awsMetadata,
	}
}

func getPublicIp() string {
	result, err := RestQuery("GET", publicIpUrl, nil, nil, nil)
	if err != nil {
		return ""
	}
	return string(result)
}

func getPublisherAgentVersion() string {
	file, err := os.Open(versionFile)
	defer file.Close()
	if err != nil {
		return ""
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return ""
	}
	return string(data)
}

func getAwsMetadata() AwsMetadata {
	// if we can't hit the AWS metadata endpoint, we must not be on AWS
	_, err := RestQuery("GET", awsMetadataUrl, nil, nil, nil)
	if err != nil {
		return AwsMetadata{
			IsAws: false,
		}
	}

	metadata := AwsMetadata{
		IsAws: true,
	}

	azResult, err := RestQuery("GET", awsMetadataUrl+"/availability-zone-id", nil, nil, nil)
	if err == nil {
		metadata.AvailabilityZoneId = string(azResult)
	}

	instanceTypeResult, err := RestQuery("GET", awsMetadataUrl+"/instance-type", nil, nil, nil)
	if err == nil {
		metadata.InstanceType = string(instanceTypeResult)
	}

	return metadata
}
