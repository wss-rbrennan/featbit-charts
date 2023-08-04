package lib

import (
	"bytes"
	"fmt"
	"strconv"
)

// * TODO: toggle api, das, els
func CreateCommandString(namespaceName string, releaseName string, apiPort int, elsPort int, dasPort int) string {
	var strBytes bytes.Buffer

	strBytes.WriteString("\nuntil (nc -vz ")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-api.")
	strBytes.WriteString(namespaceName)
	strBytes.WriteString(".svc.cluster.local ")
	strBytes.WriteString(strconv.Itoa(apiPort))
	strBytes.WriteString("); do\n    echo \"waiting for API\"; sleep 1;\ndone\n\n\nuntil (nc -vz ")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-els.")
	strBytes.WriteString(namespaceName)
	strBytes.WriteString(".svc.cluster.local ")
	strBytes.WriteString(strconv.Itoa(elsPort))
	strBytes.WriteString("); do\n    echo \"waiting for Evaluation Server\"; sleep 1;\ndone\n\n\nuntil (nc -vz ")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-das.")
	strBytes.WriteString(namespaceName)
	strBytes.WriteString(".svc.cluster.local ")
	strBytes.WriteString(strconv.Itoa(dasPort))
	strBytes.WriteString("); do\n  echo \"waiting for DA Server\"; sleep 1;\ndone\n")

	return fmt.Sprint(strBytes.String())
}

// * TODO: toggle api, das, els
func CreateWaitForInfraCommandString(namespaceName string, releaseName string, fullName string) string {
	var strBytes bytes.Buffer

	strBytes.WriteString("\nuntil (nc -vz \"")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-")
	strBytes.WriteString(fullName)
	strBytes.WriteString("-redis-master")
	strBytes.WriteString(".$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local\" 6379); do\n    echo \"waiting for Redis\"; sleep 1;\ndone\n\n\nuntil (nc -vz \"")
	strBytes.WriteString(releaseName)
	strBytes.WriteString("-")
	strBytes.WriteString(fullName)
	strBytes.WriteString("-mongodb")
	strBytes.WriteString(".$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local\" 27017); do\n    echo \"waiting for Mongodb\"; sleep 1;\ndone\n")

	return fmt.Sprint(strBytes.String())
}
