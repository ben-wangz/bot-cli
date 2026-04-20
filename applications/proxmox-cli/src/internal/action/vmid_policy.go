package action

import (
	"os"
	"strconv"
	"strings"

	"github.com/ben-wangz/bot-cli/applications/proxmox-cli/src/internal/apperr"
)

const (
	defaultOperationVMIDMin = 1001
	defaultOperationVMIDMax = 2000
	envOperationVMIDMin     = "PVE_ALLOWED_VMID_MIN"
	envOperationVMIDMax     = "PVE_ALLOWED_VMID_MAX"
)

type VMIDRange struct {
	Min int
	Max int
}

func OperationVMIDRange() (VMIDRange, error) {
	minValue, err := parseRangeBound(envOperationVMIDMin, defaultOperationVMIDMin)
	if err != nil {
		return VMIDRange{}, err
	}
	maxValue, err := parseRangeBound(envOperationVMIDMax, defaultOperationVMIDMax)
	if err != nil {
		return VMIDRange{}, err
	}
	if minValue > maxValue {
		return VMIDRange{}, apperr.New(apperr.CodeConfig, envOperationVMIDMin+" cannot be greater than "+envOperationVMIDMax)
	}
	return VMIDRange{Min: minValue, Max: maxValue}, nil
}

func EnsureOperationVMID(vmid int) error {
	r, err := OperationVMIDRange()
	if err != nil {
		return err
	}
	if vmid < r.Min || vmid > r.Max {
		return apperr.New(apperr.CodeInvalidArgs, "vmid is outside allowed operation range")
	}
	return nil
}

func parseRangeBound(envName string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(envName))
	if raw == "" {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, apperr.New(apperr.CodeConfig, envName+" must be a positive integer")
	}
	return v, nil
}
