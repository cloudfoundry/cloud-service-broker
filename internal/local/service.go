package local

import (
	"fmt"
	"log"
)

func Service(name, cachePath string) {
	instance, err := store().GetServiceInstanceDetails(nameToID(name))
	if err != nil {
		log.Fatal(err)
	}

	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker := startBroker(pakDir)
	defer broker.Stop()

	lastOperationFinalValue, _ := broker.LastOperationFinalValue(instance.GUID)

	tp := newTablePrinter("")
	serviceName, planName := lookupServiceNamesByID(broker.Client, instance.ServiceGUID, instance.PlanGUID)
	tp.row("name:", name)
	tp.row("guid:", instance.GUID)
	tp.row("offering:", serviceName)

	tp.row("plan:", planName)
	tp.row("")
	tp.row("Showing status of last operation:")
	tp.row("status:", string(lastOperationFinalValue.State))
	tp.row("message:", lastOperationFinalValue.Description)

	tp.row("")
	tp.row("Showing bindings: not implemented")

	deployment, _ := store().GetTerraformDeployment(fmt.Sprintf("tf:%s:", instance.GUID))
	tfVersion, _ := deployment.TFWorkspace().StateTFVersion()

	tp.row("")
	tp.row("Showing upgrade status: Terraform version in state is ", tfVersion.String())

	tp.print()
}
