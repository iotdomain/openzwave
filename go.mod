module github.com/iotdomain/openzwave

go 1.14

require (
	github.com/iotdomain/iotdomain-go v0.0.0-20200928060533-3e6dc24cf1bb
	github.com/jimjibone/goopenzwave v0.0.0-20180922121220-472b2577dc05
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
)

// Temporary for testing iotdomain-go
replace github.com/iotdomain/iotdomain-go => ../iotdomain-go
