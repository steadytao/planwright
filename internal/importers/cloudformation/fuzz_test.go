// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import "testing"

func FuzzImportCloudFormation(f *testing.F) {
	f.Add([]byte(`Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: planwright-fuzz-bucket
`))
	f.Add([]byte(`Resources:
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public HTTPS access.
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0
`))
	f.Add([]byte("Resources: [\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		if _, err := Import(data, "fuzz-template.yaml", FormatCloudFormation); err != nil {
			return
		}
	})
}

func FuzzImportSAM(f *testing.F) {
	f.Add([]byte(`Transform: AWS::Serverless-2016-10-31
Resources:
  Function:
    Type: AWS::Serverless::Function
    Properties:
      Runtime: provided.al2023
      Handler: bootstrap
`))
	f.Add([]byte(`Resources:
  Api:
    Type: AWS::Serverless::HttpApi
`))
	f.Add([]byte("Transform: AWS::Serverless-2016-10-31\nResources:\n  Function: [\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		if _, err := Import(data, "fuzz-sam.yaml", FormatSAM); err != nil {
			return
		}
	})
}
