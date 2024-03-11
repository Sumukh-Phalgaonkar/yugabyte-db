/*
 * Copyright (c) YugaByte, Inc.
 */

package aws

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybaclient "github.com/yugabyte/platform-go-client"
	"github.com/yugabyte/yugabyte-db/managed/yba-cli/cmd/provider/providerutil"
	"github.com/yugabyte/yugabyte-db/managed/yba-cli/cmd/util"
	ybaAuthClient "github.com/yugabyte/yugabyte-db/managed/yba-cli/internal/client"
	"github.com/yugabyte/yugabyte-db/managed/yba-cli/internal/formatter"
)

// createAWSProviderCmd represents the provider command
var createAWSProviderCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an AWS YugabyteDB Anywhere provider",
	Long:  "Create an AWS provider in YugabyteDB Anywhere",
	PreRun: func(cmd *cobra.Command, args []string) {
		providerNameFlag, err := cmd.Flags().GetString("name")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}
		if len(providerNameFlag) == 0 {
			cmd.Help()
			logrus.Fatalln(
				formatter.Colorize("No provider name found to create\n", formatter.RedColor))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		authAPI := ybaAuthClient.NewAuthAPIClientAndCustomer()

		providerName, err := cmd.Flags().GetString("name")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		allowed, version, err := authAPI.NewProviderYBAVersionCheck()
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}
		providerCode := util.AWSProviderType
		config, err := buildAWSConfig(cmd)
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		airgapInstall, err := cmd.Flags().GetBool("airgap-install")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		sshUser, err := cmd.Flags().GetString("ssh-user")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		sshPort, err := cmd.Flags().GetInt("ssh-port")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		keyPairName, err := cmd.Flags().GetString("custom-ssh-keypair-name")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		filePath, err := cmd.Flags().GetString("custom-ssh-keypair-file-path")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		var sshFileContent string
		if len(filePath) > 0 {
			sshFileContentByte, err := os.ReadFile(filePath)
			if err != nil {
				logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
			}
			sshFileContent = string(sshFileContentByte)
		}

		regions, err := cmd.Flags().GetStringArray("region")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		zones, err := cmd.Flags().GetStringArray("zone")
		if err != nil {
			logrus.Fatalf(formatter.Colorize(err.Error()+"\n", formatter.RedColor))
		}

		requestBody := ybaclient.Provider{
			Code:                 util.GetStringPointer(providerCode),
			Config:               util.StringMap(config),
			Name:                 util.GetStringPointer(providerName),
			AirGapInstall:        util.GetBoolPointer(airgapInstall),
			SshPort:              util.GetInt32Pointer(int32(sshPort)),
			SshUser:              util.GetStringPointer(sshUser),
			KeyPairName:          util.GetStringPointer(keyPairName),
			SshPrivateKeyContent: util.GetStringPointer(sshFileContent),
			Regions:              buildAWSRegions(regions, zones, allowed, version),
		}

		rCreate, response, err := authAPI.CreateProvider().
			CreateProviderRequest(requestBody).Execute()
		if err != nil {
			errMessage := util.ErrorFromHTTPResponse(response, err, "Provider: AWS", "Create")
			logrus.Fatalf(formatter.Colorize(errMessage.Error()+"\n", formatter.RedColor))
		}

		providerUUID := rCreate.GetResourceUUID()
		taskUUID := rCreate.GetTaskUUID()

		providerutil.WaitForCreateProviderTask(authAPI, providerName, providerUUID, taskUUID)
	},
}

func init() {
	createAWSProviderCmd.Flags().SortFlags = false

	// Flags needed for AWS
	createAWSProviderCmd.Flags().String("access-key-id", "",
		fmt.Sprintf("AWS Access Key ID. %s "+
			"Can also be set using environment variable %s.",
			formatter.Colorize("Required for non IAM role based providers.",
				formatter.GreenColor),
			util.AWSAccessKeyEnv))
	createAWSProviderCmd.Flags().String("secret-access-key", "",
		fmt.Sprintf("AWS Secret Access Key. %s "+
			"Can also be set using environment variable %s.",
			formatter.Colorize("Required for non IAM role based providers.",
				formatter.GreenColor),
			util.AWSSecretAccessKeyEnv))
	createAWSProviderCmd.MarkFlagsRequiredTogether("access-key-id", "secret-access-key")
	createAWSProviderCmd.Flags().Bool("use-iam-instance-profile", false,
		"[Optional] Use IAM Role from the YugabyteDB Anywhere Host. Provider "+
			"creation will fail on insufficient permissions on the host, defaults to false.")
	createAWSProviderCmd.Flags().String("hosted-zone-id", "",
		"[Optional] Hosted Zone ID corresponding to Amazon Route53.")

	createAWSProviderCmd.Flags().StringArray("region", []string{},
		"[Required] Region associated with the AWS provider. Minimum number of required "+
			"regions = 1. Provide the following comma separated fields as key-value pairs:"+
			"\"region-name=<region-name>,"+
			"vpc-id=<vpc-id>,sg-id=<security-group-id>,arch=<architecture>,yb-image=<custom-ami>\". "+
			formatter.Colorize("Region name, VPC ID and Security Group ID are required key-values.",
				formatter.GreenColor)+
			" YB Image (AMI) and Architecture (Default to x86_84, accepted in YugabyteDB "+
			"Anywhere versions >= 2.18.0) are optional. "+
			"Each region needs to be added using a separate --region flag. "+
			"Example: --region region-name=us-west-2,vpc-id=<vpc-id>,sg-id=<security-group> "+
			"--region region-name=us-east-2,vpc-id=<vpc-id>,sg-id=<security-group>")
	createAWSProviderCmd.Flags().StringArray("zone", []string{},
		"[Required] Zone associated to the AWS Region defined. "+
			"Provide the following comma separated fields as key-value pairs:"+
			"\"zone-name=<zone-name>,region-name=<region-name>,subnet=<subnet-id>,"+
			"secondary-subnet=<secondary-subnet-id>\"."+
			formatter.Colorize("Zone name, Region name and subnet IDs are required values. ",
				formatter.GreenColor)+
			"Secondary subnet ID is optional. Each --region definition "+
			"must have atleast one corresponding --zone definition. Multiple --zone definitions "+
			"can be provided per region."+
			"Each zone needs to be added using a separate --zone flag. "+
			"Example: --zone zone-name=us-west-2a,region-name=us-west-2,subnet=<subnet-id>"+
			" --zone zone-name=us-west-2b,region-name=us-west-2,subnet=<subnet-id>")

	createAWSProviderCmd.Flags().String("ssh-user", "",
		"[Optional] SSH User to access the YugabyteDB nodes.")
	createAWSProviderCmd.Flags().Int("ssh-port", 22,
		"[Optional] SSH Port to access the YugabyteDB nodes.")
	createAWSProviderCmd.Flags().String("custom-ssh-keypair-name", "",
		"[Optional] Provide custom key pair name to access YugabyteDB nodes. "+
			"YugabyteDB Anywhere will generate key pairs to access YugabyteDB nodes.")
	createAWSProviderCmd.Flags().String("custom-ssh-keypair-file-path", "",
		fmt.Sprintf("[Optional] Provide custom key pair file path to access YugabyteDB nodes. %s",
			formatter.Colorize("Required with --custom-ssh-keypair-name.",
				formatter.GreenColor)))
	createAWSProviderCmd.MarkFlagsRequiredTogether("custom-ssh-keypair-name",
		"custom-ssh-keypair-file-path")

	createAWSProviderCmd.Flags().Bool("airgap-install", false,
		"[Optional] Are YugabyteDB nodes installed in an air-gapped environment,"+
			" lacking access to the public internet for package downloads, "+
			"defaults to false.")

}

func buildAWSConfig(cmd *cobra.Command) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	isIAM, err := cmd.Flags().GetBool("use-iam-instance-profile")
	if err != nil {
		return nil, err
	}
	hostedZoneID, err := cmd.Flags().GetString("hosted-zone-id")
	if err != nil {
		return nil, err
	}
	if len(hostedZoneID) > 0 {
		config["HOSTED_ZONE_ID"] = hostedZoneID
	}

	if !isIAM {
		accessKeyID, err := cmd.Flags().GetString("access-key-id")
		if err != nil {
			return nil, err
		}
		secretAccessKey, err := cmd.Flags().GetString("secret-access-key")
		if err != nil {
			return nil, err
		}
		if len(accessKeyID) == 0 && len(secretAccessKey) == 0 {
			awsCreds, err := util.AwsCredentialsFromEnv()
			if err != nil {
				return nil, err
			}
			config[util.AWSAccessKeyEnv] = awsCreds.AccessKeyID
			config[util.AWSSecretAccessKeyEnv] = awsCreds.SecretAccessKey
		} else {
			config[util.AWSAccessKeyEnv] = accessKeyID
			config[util.AWSSecretAccessKeyEnv] = secretAccessKey
		}
	}
	return config, nil
}

func buildAWSRegions(regionStrings, zoneStrings []string, allowed bool,
	version string) (res []ybaclient.Region) {
	if len(regionStrings) == 0 {
		logrus.Fatalln(
			formatter.Colorize("Atleast one region is required per provider.",
				formatter.RedColor))
	}
	for _, regionString := range regionStrings {
		region := map[string]string{}
		for _, regionInfo := range strings.Split(regionString, ",") {
			kvp := strings.Split(regionInfo, "=")
			if len(kvp) != 2 {
				logrus.Fatalln(
					formatter.Colorize("Incorrect format in region description.",
						formatter.RedColor))
			}
			key := kvp[0]
			val := kvp[1]
			switch key {
			case "region-name":
				if len(strings.TrimSpace(val)) != 0 {
					region["name"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "vpc-id":
				if len(strings.TrimSpace(val)) != 0 {
					region["vpc-id"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "sg-id":
				if len(strings.TrimSpace(val)) != 0 {
					region["sg-id"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "arch":
				if len(strings.TrimSpace(val)) != 0 {
					region["arch"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "yb-image":
				if len(strings.TrimSpace(val)) != 0 {
					region["yb-image"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			}
		}
		if _, ok := region["name"]; !ok {
			logrus.Fatalln(
				formatter.Colorize("Name not specified in region.",
					formatter.RedColor))
		}
		if _, ok := region["vpc-id"]; !ok {
			logrus.Fatalln(
				formatter.Colorize("VPC ID not specified in region info.",
					formatter.RedColor))

		}
		if _, ok := region["sg-id"]; !ok {
			logrus.Fatalln(
				formatter.Colorize("Security Group ID not specified in region info.",
					formatter.RedColor))
		}

		zones := buildAWSZones(zoneStrings, region["name"])
		r := ybaclient.Region{
			Code:            util.GetStringPointer(region["name"]),
			Name:            util.GetStringPointer(region["name"]),
			SecurityGroupId: util.GetStringPointer(region["sg-id"]),
			VnetName:        util.GetStringPointer(region["vpc-id"]),
			YbImage:         util.GetStringPointer(region["yb-image"]),
			Zones:           zones,
		}
		if allowed {
			r.Details = &ybaclient.RegionDetails{
				CloudInfo: &ybaclient.RegionCloudInfo{
					Aws: &ybaclient.AWSRegionCloudInfo{
						YbImage: util.GetStringPointer(region["yb-image"]),
						Arch:    util.GetStringPointer(region["arch"]),
					},
				},
			}
		} else {
			logrus.Info(
				fmt.Sprintf("YugabyteDB Anywhere version %s does not support specifying "+
					"Architecture, ignoring value.", version))
		}
		res = append(res, r)
	}
	return res
}

func buildAWSZones(zoneStrings []string, regionName string) (res []ybaclient.AvailabilityZone) {
	for _, zoneString := range zoneStrings {
		zone := map[string]string{}
		for _, zoneInfo := range strings.Split(zoneString, ",") {
			kvp := strings.Split(zoneInfo, "=")
			if len(kvp) != 2 {
				logrus.Fatalln(
					formatter.Colorize("Incorrect format in zone description",
						formatter.RedColor))
			}
			key := kvp[0]
			val := kvp[1]
			switch key {
			case "zone-name":
				if len(strings.TrimSpace(val)) != 0 {
					zone["name"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "region-name":
				if len(strings.TrimSpace(val)) != 0 {
					zone["region-name"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "subnet":
				if len(strings.TrimSpace(val)) != 0 {
					zone["subnet"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			case "secondary-subnet":
				if len(strings.TrimSpace(val)) != 0 {
					zone["secondary-subnet"] = val
				} else {
					providerutil.ValueNotFoundForKeyError(key)
				}
			}
		}
		if _, ok := zone["name"]; !ok {
			logrus.Fatalln(
				formatter.Colorize("Name not specified in zone.",
					formatter.RedColor))
		}
		if _, ok := zone["region-name"]; !ok {
			logrus.Fatalln(
				formatter.Colorize("Region name not specified in zone.",
					formatter.RedColor))
		}
		if _, ok := zone["subnet"]; !ok {
			logrus.Fatalln(
				formatter.Colorize("Subnet not specified in zone info.",
					formatter.RedColor))
		}

		if strings.Compare(zone["region-name"], regionName) == 0 {
			z := ybaclient.AvailabilityZone{
				Code:            util.GetStringPointer(zone["name"]),
				Name:            zone["name"],
				SecondarySubnet: util.GetStringPointer(zone["secondary-subnet"]),
				Subnet:          util.GetStringPointer(zone["subnet"]),
			}
			res = append(res, z)
		}
	}
	if len(res) == 0 {
		logrus.Fatalln(
			formatter.Colorize("Atleast one zone is required per region.",
				formatter.RedColor))
	}
	return res
}
