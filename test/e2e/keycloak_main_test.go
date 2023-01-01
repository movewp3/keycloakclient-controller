package e2e

import (
	"testing"
)

type deployedOperatorTestStep struct {
	prepareTestEnvironmentSteps []environmentInitializationStep
	testFunction                func(*testing.T, string) error
}

type environmentInitializationStep func(*testing.T, string) error

type CRDTestStruct struct {
	prepareEnvironmentSteps []environmentInitializationStep
	testSteps               map[string]deployedOperatorTestStep
}

func TestKeycloakCRDS(t *testing.T) {
}

func runTestsFromCRDInterface(t *testing.T, crd *CRDTestStruct) {
}
