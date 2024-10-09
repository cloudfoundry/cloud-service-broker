// Copyright 2019 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbservice

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func ExampleUseVcapServices() {
	_ = os.Setenv("VCAP_SERVICES", `{
  "p.mysql": [
    {
      "label": "p.mysql",
      "name": "my-instance",
      "plan": "db-medium",
      "provider": null,
      "syslog_drain_url": null,
      "tags": [
        "mysql"
      ],
      "credentials": {
        "hostname": "10.0.0.20",
        "jdbcUrl": "jdbc:mysql://10.0.0.20:3306/service_instance_db?user=fefcbe8360854a18a7994b870e7b0bf5\u0026password=z9z6eskdbs1rhtxt",
        "name": "service_instance_db",
        "password": "z9z6eskdbs1rhtxt",
        "port": 3306,
        "uri": "mysql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true",
        "username": "fefcbe8360854a18a7994b870e7b0bf5"
      },
      "volume_mounts": []
    }
  ]
}`)
	_ = UseVcapServices()
	fmt.Println(viper.Get(dbHostProp))
	fmt.Println(viper.Get(dbUserProp))
	fmt.Println(viper.Get(dbPassProp))
	fmt.Println(viper.Get(dbNameProp))

	// Output:
	// 10.0.0.20
	// fefcbe8360854a18a7994b870e7b0bf5
	// z9z6eskdbs1rhtxt
	// service_instance_db
}

var _ = Describe("VCAP", func() {
	Describe("parsing VCAP_SERVICES", func() {
		It("fails when VCAP_SERVICES is empty", func() {
			_, err := ParseVcapServices("")
			Expect(err).To(MatchError("error unmarshalling VCAP_SERVICES: unexpected end of JSON input"))
		})

		It("parses VCAP_SERVICES for Google Cloud", func() {
			_, err := ParseVcapServices(`{
  "google-cloudsql-mysql": [
    {
      "binding_name": "testbinding",
      "instance_name": "testinstance",
      "name": "kf-binding-tt2-mystorage",
      "label": "google-storage",
      "tags": [
        "gcp",
        "cloudsql",
        "mysql"
      ],
      "plan": "nearline",
      "credentials": {
	"CaCert": "-truncated-",
	"ClientCert": "-truncated-",
	"ClientKey": "-truncated-",
	"Email": "pcf-binding-testbind@test-gsb.iam.gserviceaccount.com",
        "Name": "pcf-binding-testbind",
        "Password": "PASSWORD",
        "PrivateKeyData": "PRIVATEKEY",
        "ProjectId": "test-gsb",
        "Sha1Fingerprint": "aa3bade266136f733642ebdb4992b89eb05f83c4",
        "UniqueId": "108868434450972082663",
        "UriPrefix": "",
        "Username": "newuseraccount",
        "database_name": "service_broker",
        "host": "127.0.0.1",
        "instance_name": "pcf-sb-1-1561406852899716453",
        "last_master_operation_id": "",
        "region": "",
        "uri": "mysql://newuseraccount:PASSWORD@127.0.0.1/service_broker?ssl_mode=required"
      }
    }
  ]
}`)
			Expect(err).NotTo(HaveOccurred())
		})

		It("parses VCAP_SERVICES for 'p.mysql' tile", func() {
			_, err := ParseVcapServices(`{
  "p.mysql": [
    {
      "label": "p.mysql",
      "name": "my-instance",
      "plan": "db-medium",
      "provider": null,
      "syslog_drain_url": null,
      "tags": [
        "mysql"
      ],
      "credentials": {
        "hostname": "10.0.0.20",
        "jdbcUrl": "jdbc:mysql://10.0.0.20:3306/service_instance_db?user=fefcbe8360854a18a7994b870e7b0bf5\u0026password=z9z6eskdbs1rhtxt",
        "name": "service_instance_db",
        "password": "z9z6eskdbs1rhtxt",
        "port": 3306,
        "uri": "mysql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true",
        "username": "fefcbe8360854a18a7994b870e7b0bf5"
      },
      "volume_mounts": []
    }
  ]
}`)
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails when VCAP_SERVICES has more than one MySQL tag", func() {
			_, err := ParseVcapServices(`{
  "google-cloudsql-mysql": [
    {
      "binding_name": "testbinding",
      "instance_name": "testinstance",
      "name": "kf-binding-tt2-mystorage",
      "label": "google-storage",
      "tags": [
        "gcp",
        "cloudsql",
        "mysql"
      ],
      "plan": "nearline",
      "credentials": {
	"CaCert": "-truncated-",
	"ClientCert": "-truncated-",
	"ClientKey": "-truncated-",
	"Email": "pcf-binding-testbind@test-gsb.iam.gserviceaccount.com",
        "Name": "pcf-binding-testbind",
        "Password": "PASSWORD",
        "PrivateKeyData": "PRIVATEKEY",
        "ProjectId": "test-gsb",
        "Sha1Fingerprint": "aa3bade266136f733642ebdb4992b89eb05f83c4",
        "UniqueId": "108868434450972082663",
        "UriPrefix": "",
        "Username": "newuseraccount",
        "database_name": "service_broker",
        "host": "127.0.0.1",
        "instance_name": "pcf-sb-1-1561406852899716453",
        "last_master_operation_id": "",
        "region": "",
        "uri": "mysql://newuseraccount:PASSWORD@127.0.0.1/service_broker?ssl_mode=required"
      }
    },
    {
      "label": "p.mysql",
      "name": "my-instance",
      "plan": "db-medium",
      "provider": null,
      "syslog_drain_url": null,
      "tags": [
        "mysql"
      ],
      "credentials": {
        "hostname": "10.0.0.20",
        "jdbcUrl": "jdbc:mysql://10.0.0.20:3306/service_instance_db?user=fefcbe8360854a18a7994b870e7b0bf5\u0026password=z9z6eskdbs1rhtxt",
        "name": "service_instance_db",
        "password": "z9z6eskdbs1rhtxt",
        "port": 3306,
        "uri": "mysql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true",
        "username": "fefcbe8360854a18a7994b870e7b0bf5"
      },
      "volume_mounts": []
    }
  ]
}`)
			Expect(err).To(MatchError("error finding MySQL tag: the variable VCAP_SERVICES must have one VCAP service with a tag of 'mysql'. There are currently 2 VCAP services with the tag 'mysql'"))
		})

		It("fails when VCAP_SERVICES has zero MySQL tag", func() {
			_, err := ParseVcapServices(`{
  "p.mysql": [
    {
      "label": "p.mysql",
      "name": "my-instance",
      "plan": "db-medium",
      "provider": null,
      "syslog_drain_url": null,
      "tags": [
        "notmysql"
      ],
      "credentials": {
        "hostname": "10.0.0.20",
        "jdbcUrl": "jdbc:mysql://10.0.0.20:3306/service_instance_db?user=fefcbe8360854a18a7994b870e7b0bf5\u0026password=z9z6eskdbs1rhtxt",
        "name": "service_instance_db",
        "password": "z9z6eskdbs1rhtxt",
        "port": 3306,
        "uri": "mysql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true",
        "username": "fefcbe8360854a18a7994b870e7b0bf5"
      },
      "volume_mounts": []
    }
  ]
}
`)
			Expect(err).To(MatchError("error finding MySQL tag: the variable VCAP_SERVICES must have one VCAP service with a tag of 'mysql'. There are currently 0 VCAP services with the tag 'mysql'"))
		})

		It("succeeds when VCAP_SERVICES has multiple lists but only one with a MySQL tag", func() {
			_, err := ParseVcapServices(`{
  "google-cloudsql-mysql": [
    {
      "binding_name": "testbinding",
      "instance_name": "testinstance",
      "name": "kf-binding-tt2-mystorage",
      "label": "google-storage",
      "tags": [
        "gcp",
        "cloudsql"
      ],
      "plan": "nearline",
      "credentials": {
				"CaCert": "-truncated-",
				"ClientCert": "-truncated-",
				"ClientKey": "-truncated-",
				"Email": "pcf-binding-testbind@test-gsb.iam.gserviceaccount.com",
        "Name": "pcf-binding-testbind",
        "Password": "PASSWORD",
        "PrivateKeyData": "PRIVATEKEY",
        "ProjectId": "test-gsb",
        "Sha1Fingerprint": "aa3bade266136f733642ebdb4992b89eb05f83c4",
        "UniqueId": "108868434450972082663",
        "UriPrefix": "",
        "Username": "newuseraccount",
        "database_name": "service_broker",
        "host": "127.0.0.1",
        "instance_name": "pcf-sb-1-1561406852899716453",
        "last_master_operation_id": "",
        "region": "",
        "uri": "mysql://newuseraccount:PASSWORD@127.0.0.1/service_broker?ssl_mode=required"
      }
    }
  ],
  "user-provided": [
  	{
      "binding_name": "testbinding",
      "instance_name": "testinstance",
      "name": "kf-binding-tt2-mystorage",
      "label": "google-storage",
      "tags": [
        "gcp",
        "cloudsql",
        "mysql"
      ],
      "plan": "nearline",
      "credentials": {
				"CaCert": "-truncated-",
				"ClientCert": "-truncated-",
				"ClientKey": "-truncated-",
				"Email": "pcf-binding-testbind@test-gsb.iam.gserviceaccount.com",
        "Name": "pcf-binding-testbind",
        "Password": "PASSWORD",
        "PrivateKeyData": "PRIVATEKEY",
        "ProjectId": "test-gsb",
        "Sha1Fingerprint": "aa3bade266136f733642ebdb4992b89eb05f83c4",
        "UniqueId": "108868434450972082663",
        "UriPrefix": "",
        "Username": "newuseraccount",
        "database_name": "service_broker",
        "host": "127.0.0.1",
        "instance_name": "pcf-sb-1-1561406852899716453",
        "last_master_operation_id": "",
        "region": "",
        "uri": "mysql://newuseraccount:PASSWORD@127.0.0.1/service_broker?ssl_mode=required"
      }
    }
  ]
}`)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("setting database credentials", func() {
		It("succeeds when the parsed VCAP_SERVICES is empty", func() {
			Expect(SetDatabaseCredentials(VcapService{})).To(Succeed())
		})

		It("succeeds when the parsed VCAP_SERVICES represents a valid Google Cloud service", func() {
			vs := VcapService{
				BindingName:  "testbinding",
				InstanceName: "testinstance",
				Name:         "kf-binding-tt2-mystorage",
				Label:        "google-storage",
				Tags:         []string{"gcp", "cloudsql", "mysql"},
				Plan:         "nearline",
				Credentials: map[string]any{
					"CaCert":                   "-truncated-",
					"ClientCert":               "-truncated-",
					"ClientKey":                "-truncated-",
					"Email":                    "pcf-binding-testbind@test-gsb.iam.gserviceaccount.com",
					"Name":                     "pcf-binding-testbind",
					"Password":                 "Ukd7QEmrfC7xMRqNmTzHCbNnmBtNceys1olOzLoSm4k",
					"PrivateKeyData":           "-truncated-",
					"ProjectId":                "test-gsb",
					"Sha1Fingerprint":          "aa3bade266136f733642ebdb4992b89eb05f83c4",
					"UniqueId":                 "108868434450972082663",
					"UriPrefix":                "",
					"Username":                 "newuseraccount",
					"database_name":            "service_broker",
					"host":                     "104.154.90.3",
					"instance_name":            "pcf-sb-1-1561406852899716453",
					"last_master_operation_id": "",
					"region":                   "",
					"uri":                      "mysql://newuseraccount:Ukd7QEmrfC7xMRqNmTzHCbNnmBtNceys1olOzLoSm4k@104.154.90.3/service_broker?ssl_mode=required",
				},
			}
			Expect(SetDatabaseCredentials(vs)).To(Succeed())
		})

		It("succeeds when the parsed VCAP_SERVICES represents a valid 'p.mysql' tile service", func() {
			vs := VcapService{
				BindingName:  "",
				InstanceName: "",
				Name:         "my-instance",
				Label:        "p.mysql",
				Tags:         []string{"mysql"},
				Plan:         "db-medium",
				Credentials: map[string]any{
					"hostname": "10.0.0.20",
					"jdbcUrl":  "jdbc:mysql://10.0.0.20:3306/service_instance_db?user=fefcbe8360854a18a7994b870e7b0bf5&password=z9z6eskdbs1rhtxt",
					"name":     "service_instance_db",
					"password": "z9z6eskdbs1rhtxt",
					"port":     3306,
					"uri":      "mysql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true",
					"username": "fefcbe8360854a18a7994b870e7b0bf5",
				},
			}
			Expect(SetDatabaseCredentials(vs)).To(Succeed())
		})

		It("fails when there is an malformed URI", func() {
			vs := VcapService{
				BindingName:  "",
				InstanceName: "",
				Name:         "my-instance",
				Label:        "p.mysql",
				Tags:         []string{"mysql"},
				Plan:         "db-medium",
				Credentials: map[string]any{
					"hostname": "10.0.0.20",
					"jdbcUrl":  "jdbc:mysql://10.0.0.20:3306/service_instance_db?user=fefcbe8360854a18a7994b870e7b0bf5&password=z9z6eskdbs1rhtxt",
					"name":     "service_instance_db",
					"password": "z9z6eskdbs1rhtxt",
					"port":     3306,
					"uri":      "mys@!ql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true",
					"username": "fefcbe8360854a18a7994b870e7b0bf5",
				},
			}
			Expect(SetDatabaseCredentials(vs)).To(MatchError(`error parsing credentials uri field: parse "mys@!ql://fefcbe8360854a18a7994b870e7b0bf5:z9z6eskdbs1rhtxt@10.0.0.20:3306/service_instance_db?reconnect=true": first path segment in URL cannot contain colon`))
		})
	})
})
