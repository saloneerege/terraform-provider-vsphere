package vsphere

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"gitlab.eng.vmware.com/openapi-sdks/vmware-infrastructure-sdk-go/runtime"
	"gitlab.eng.vmware.com/openapi-sdks/vmware-infrastructure-sdk-go/services/vsphere/vcenter/namespaces"
)

const NAMESPACE_RESOURCE = "namespace1"

func TestAccResourceVsphereNamespace_basic(t *testing.T) {
	nameSpace := "terraform_" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			runAPISweeper()
			testAccPreCheck(t)
			testAccResourcevSphereNamespacePreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceNameSpaceCheckExists(nameSpace),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereNamespaceConfigBasic(nameSpace),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceNameSpaceCheckExists(nameSpace),
					resource.TestCheckResourceAttrSet("vsphere_namespace."+NAMESPACE_RESOURCE, "namespace"),
				),
			},
		},
	})
}

func testAccResourcevSphereNamespacePreCheck(t *testing.T) {
	fmt.Println("testAccResourcevSphereNamespacePreCheck")
	if os.Getenv("TF_VAR_VSPHERE_NAMESPACE_CLUSTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAMESPACE_CLUSTER to run vsphere_namespace acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_SUBJECT") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_SUBJECT to run vsphere_namespace acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_SUBJECT_TYPE") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_SUBJECT_TYPE  to run vsphere_namespace acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_DOMAIN") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_DOMAIN  to run vsphere_namespace acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_ROLE") == "" {
		t.Skip("set TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_ROLE  to run vsphere_namespace acceptance tests")
	}
	fmt.Println(" end of testAccResourcevSphereNamespacePreCheck")
}

func testAccResourceNameSpaceCheckExists(nameSpace string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[nameSpace]
		if !ok {
			return fmt.Errorf("not found: %s", nameSpace)
		}
		nameSpaceID := rs.Primary.Attributes["namespace"]
		fmt.Println("testAccResourceNameSpaceCheckExists")
		APIClient := testAccProvider.Meta().(*VSphereClient).apiClient
		fmt.Println(APIClient.SessionID)
		namespaceConfig := GetNamespaceRuntimeConfig(APIClient.SessionID, APIClient.BasePath, APIClient.InsecureFlag)

		ctx := context.WithValue(context.Background(), runtime.ContextAPIKey, runtime.APIKey{
			Key:    APIClient.SessionID,
			Prefix: "",
		})

		client := namespaces.NewAPIClient(namespaceConfig)
		_, _, err := client.InstancesApi.Get(ctx, nameSpaceID)
		if err != nil {
			return fmt.Errorf("error while reading namespace :%v", err)
		}

		fmt.Printf("Namespace  %s created successfully \n", nameSpace)

		return nil
	}
}

func testAccResourceVsphereNamespaceConfigBasic(nameSpace string) string {
	return fmt.Sprintf(`
   resource vsphere_namespace "%s" {
   namespace = "%s"
   cluster = "%s"
   access_list {
     subject = "%s"
     subject_type = "%s"
     domain = "%s"
     role = "%s"
   }
 }
`, NAMESPACE_RESOURCE,
		nameSpace,
		os.Getenv("TF_VAR_VSPHERE_NAMESPACE_CLUSTER"),
		os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_SUBJECT"),
		os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_SUBJECT_TYPE"),
		os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_DOMAIN"),
		os.Getenv("TF_VAR_VSPHERE_NAMESPACE_ACCESSLIST_ROLE"),
	)
}

func sweepAPIClient() (*APISessionClient, error) {
	config := Config{
		InsecureFlag:  true,
		Persist:       false,
		User:          os.Getenv("TF_VAR_VSPHERE_USER"),
		Password:      os.Getenv("TF_VAR_VSPHERE_PASSWORD"),
		VSphereServer: os.Getenv("TF_VAR_VSPHERE_SERVER"),
	}
	return config.ApiSessionClient()
}
