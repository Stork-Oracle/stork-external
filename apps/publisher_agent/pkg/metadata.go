package publisher_agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer"
	"github.com/rs/zerolog"
)

const (
	publicIpUrl    = "https://api.ipify.org"
	awsMetadataUrl = "http://169.254.169.254/latest/meta-data"
	versionFile    = "version.txt"
)

type PublisherMetadata struct {
	PublisherKey          shared.PublisherKey       `json:"publisher_key"`
	SignatureType         shared.SignatureType      `json:"signature_type"`
	PublisherAgentVersion string                    `json:"publisher_agent_version"`
	Architecture          string                    `json:"architecture"`
	PublicIp              string                    `json:"public_ip"`
	AwsMetadata           AwsMetadata               `json:"aws_metadata"`
	Config                StorkPublisherAgentConfig `json:"config"`
}

type AwsMetadata struct {
	IsAws              bool   `json:"is_aws"`
	InstanceType       string `json:"instance_type"`
	AvailabilityZoneId string `json:"availability_zone_id"`
}

type PublisherMetadataReporter struct {
	publicKey                shared.PublisherKey
	signatureType            shared.SignatureType
	reportPeriod             time.Duration
	publisherMetadataBaseUrl string
	storkAuthSigner          signer.StorkAuthSigner
	logger                   zerolog.Logger
	config                   StorkPublisherAgentConfig
}

func NewPublisherMetadataReporter(
	publicKey shared.PublisherKey,
	signatureType shared.SignatureType,
	reportPeriod time.Duration,
	publisherMetadataBaseUrl string,
	storkAuthSigner signer.StorkAuthSigner,
	logger zerolog.Logger,
	config StorkPublisherAgentConfig,
) *PublisherMetadataReporter {
	return &PublisherMetadataReporter{
		publicKey:                publicKey,
		signatureType:            signatureType,
		reportPeriod:             reportPeriod,
		publisherMetadataBaseUrl: publisherMetadataBaseUrl,
		storkAuthSigner:          storkAuthSigner,
		logger:                   logger,
		config:                   config,
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
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshaling publisher metadata: %v", err)
	}
	authHeaders, err := p.storkAuthSigner.GetAuthHeaders()
	if err != nil {
		return fmt.Errorf("error getting auth headers: %v", err)
	}
	_, err = RestQuery(
		"POST",
		p.publisherMetadataBaseUrl+"/v1/publisher/metadata",
		nil,
		bytes.NewReader(metadataJson),
		authHeaders,
	)
	if err != nil {
		return fmt.Errorf("error reporting publisher metadata: %v", err)
	}

	p.logger.Info().Msgf("Reported publisher metadata: %v", metadata)
	return nil
}

func (p *PublisherMetadataReporter) getMetadata() PublisherMetadata {
	awsMetadata := getAwsMetadata()
	architecture := runtime.GOARCH
	publicIp := getPublicIp()
	version := getPublisherAgentVersion()

	return PublisherMetadata{
		PublisherKey:          p.publicKey,
		SignatureType:         p.signatureType,
		PublisherAgentVersion: version,
		Architecture:          architecture,
		PublicIp:              publicIp,
		AwsMetadata:           awsMetadata,
		Config:                p.config,
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

	azResult, err := RestQuery("GET", awsMetadataUrl+"/placement/availability-zone-id", nil, nil, nil)
	if err == nil {
		metadata.AvailabilityZoneId = string(azResult)
	}

	instanceTypeResult, err := RestQuery("GET", awsMetadataUrl+"/instance-type", nil, nil, nil)
	if err == nil {
		metadata.InstanceType = string(instanceTypeResult)
	}

	return metadata
}
