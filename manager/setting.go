package manager

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/longhorn-manager/types"
	"github.com/rancher/longhorn-manager/util"

	longhorn "github.com/rancher/longhorn-manager/k8s/pkg/apis/longhorn/v1alpha1"
)

func (m *VolumeManager) GetSettingValueExisted(sName types.SettingName) (string, error) {
	setting, err := m.GetSetting(sName)
	if err != nil {
		return "", err
	}
	if setting.Value == "" {
		return "", fmt.Errorf("setting %v is empty", sName)
	}
	return setting.Value, nil
}

func (m *VolumeManager) GetSetting(sName types.SettingName) (*longhorn.Setting, error) {
	return m.ds.GetSetting(sName)
}

func (m *VolumeManager) ListSettings() (map[types.SettingName]*longhorn.Setting, error) {
	return m.ds.ListSettings()
}

func (m *VolumeManager) ListSettingsSorted() ([]*longhorn.Setting, error) {
	settingMap, err := m.ListSettings()
	if err != nil {
		return []*longhorn.Setting{}, err
	}

	settings := make([]*longhorn.Setting, len(settingMap))
	settingNames, err := sortKeys(settingMap)
	if err != nil {
		return []*longhorn.Setting{}, err
	}
	for i, settingName := range settingNames {
		settings[i] = settingMap[types.SettingName(settingName)]
	}
	return settings, nil
}

func (m *VolumeManager) CreateOrUpdateSetting(s *longhorn.Setting) (*longhorn.Setting, error) {
	err := m.SettingValidation(s.Name, s.Value)
	if err != nil {
		return nil, err
	}
	setting, err := m.ds.UpdateSetting(s)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return m.ds.CreateSetting(s)
		}
		return nil, err
	}
	return setting, nil
}

func (m *VolumeManager) SettingValidation(name, value string) error {
	sName := types.SettingName(name)

	switch sName {
	case types.SettingNameBackupTarget:
		// additional check whether have $ or , have been set in BackupTarget
		regStr := `[\$\,]`
		reg := regexp.MustCompile(regStr)
		findStr := reg.FindAllString(value, -1)
		if len(findStr) != 0 {
			return fmt.Errorf("fail to set settings with invalid BackupTarget %s, contains %v", value, strings.Join(findStr, " or "))
		}
	case types.SettingNameStorageOverProvisioningPercentage:
		// additional check whether over provisioning percentage is positive
		value, err := util.ConvertSize(value)
		if err != nil || value < 0 {
			return fmt.Errorf("fail to set settings with invalid StorageOverProvisioningPercentag %v, value should be positive", value)
		}
	case types.SettingNameStorageMinimalAvailablePercentage:
		// additional check whether minimal available percentage is between 0 to 100
		value, err := util.ConvertSize(value)
		if err != nil || value < 0 || value > 100 {
			return fmt.Errorf("fail to set settings with invalid StorageMinimalAvailablePercentage %v, value should between 0 to 100", value)
		}
	case types.SettingNameDefaultReplicaCount:
		c, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("fail to set setting with invalid %s: %v", types.SettingDefinitions[types.SettingNameDefaultEngineImage].DisplayName, err)
		}
		if c < 1 || c > 20 {
			return fmt.Errorf("fail to set setting with invalid %s: valid value must between 1 to 20",
				types.SettingDefinitions[types.SettingNameDefaultEngineImage].DisplayName)
		}
	}
	return nil
}
