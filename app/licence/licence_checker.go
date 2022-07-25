package licence_checker

import (
	"encoding/base64"
	"os"

	licence "github.com/IgorPidik/go-jwt-licence"
	"github.com/golang-jwt/jwt/v4"
)

type LicenceChecker struct {
	LicenceValidator *licence.LicenceValidator
	LicenceFilePath  string
}

func NewLicenceChecker(licenceFilePath string) (*LicenceChecker, error) {
	key := `LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQ0lqQU5CZ2txaGtpRzl3MEJB
UUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUF1RERtVm84R2dOTC8wQ2d3RFBFeQovVzRj
bWQ5Y0cxMnN6UXdvbGp4TG9IbWhZMGJPUjV1MmpicjVQK0QzbnBQajRzYURZdGd5
WDhuYWZUQzJabk1GCjA4QUdwVVBrbWUxbDVLVUY5cFprei9hNlUzWlBNOHJMRnNi
cXdueTBBeFFoYVREZFRUY1dNMTFmaWdaNEN0d2sKdGhFMzZ1RjFYb0dlQURnYkdI
dU5QbVh5alVrWktHaUxIc1YvQWliYlAxL2xUbE9DWCtjMCsxRTdjOXV3L0I4TQpx
b0x2dHZSVWFmNGxnYjhPNmhlVjZjTUFEaFhJT1UyNmxVaWFtZnEvbzlDV0xSOHpo
Nk0rdW9CQytFbVZhcU5lClZQZlFRV2xGb1FOT3ZQU203ZEtBK015ZjdRbndoSmM1
S0hIM1dSd3lGL2N4TVRGTnNkbGM0M09uSzJtYW9FQUkKMGJlTDFDc1p1RjhqVjlq
Nyt2TFNrTVg0OUlmVWY2U1lBeTl1ajRIOHpjeTBISlNOZC80dEpqR256b0w3OVNG
cgo5QkNlcGl5SDJOVzRoTXh0aXUwK3Bmc3RqV2lnK1N5bXJjS1hLSTFBL05DVFMx
clNDblNDcUNIWWdvZG9IS2F0CkIxcVZsRElTU0tRZ1IraU9xVzd2K205cDFvakVJ
TXR2K0pCMXRyb0s4TlhrblRlL1FoQmF0Z0VOZDRjTGRabWMKZDdVZDlHSWQwb0M1
c3pIY2orZFkrT0piZVd1VFV4RmZIeTViQ0czTlhqZSs3Uk9qZlNySE1INzVzMmpL
Yi9GRQpCYTBKc0dqR2J3dU5NTlRQN2lNaVRpbkU1cnBqbDM4MXFuNytzSlhLV25Y
NjltMzB3T1NxYWNLTGRBd2Vvb3c1CmVPUzA1N0xCbUxrY29ndjE3S25WYncwQ0F3
RUFBUT09Ci0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLQo=`

	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		// TODO: replace with custom error for security reasons
		return nil, err
	}

	publicKey, parseErr := jwt.ParseRSAPublicKeyFromPEM(decodedKey)
	if parseErr != nil {
		return nil, parseErr
	}

	validator := licence.NewLicenceValidatorFromPublicKey(publicKey)
	return &LicenceChecker{
		LicenceValidator: validator,
		LicenceFilePath:  licenceFilePath,
	}, nil
}

func (lc *LicenceChecker) CheckLicence() (*licence.LicenceData, error) {
	licence, fileErr := os.ReadFile(lc.LicenceFilePath)
	if fileErr != nil {
		return nil, fileErr
	}

	return lc.LicenceValidator.ValidateLicence(string(licence))
}
