package utils

import "github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

func WorkerPresentInMachineSets(machineSets []mcaasapi.MachineSet, workername string) bool {
	for _, ms := range machineSets {
		if ms.Name == workername {
			return true
		}
	}
	return false
}

func RemoveWorkerFromMachineSets(machineSets []mcaasapi.MachineSet, workername string) []mcaasapi.MachineSet {
	for i, ms := range machineSets {
		if ms.Name == workername {
			return append(machineSets[:i], machineSets[i+1:]...)
		}
	}
	return machineSets
}
