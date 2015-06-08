package collectors

import (
	"os"
	"time"

	"bosun.org/metadata"
	"bosun.org/opentsdb"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_dsc_mof, Interval: time.Minute * 5})
	collectors = append(collectors, &IntervalCollector{F: c_dsc_status, Interval: time.Minute * 5})
}

const (
	dscLCM = "dsc.lcm."
	dscMof = "dsc.mof."
)

var (
	dscpath     = os.ExpandEnv(`${SYSTEMROOT}\system32\Configuration\`)
	mapMofFiles = map[string]string{
		"MetaConfig.mof":       "Meta_Config",
		"Current.mof":          "Current_Config",
		"backup.mof":           "Backup_Config",
		"pending.mof":          "Pending_Config",
		"DSCStatusHistory.mof": "DSC_History",
		"DSCEngineCache.mof":   "DSC_Cache",
	}
)

// c_dsc_mof monitors the size and last modified time of each mof file.
// These out of band metrics can be used to verify the DSC WMI Status metrics.
func c_dsc_mof() (opentsdb.MultiDataPoint, error) {
	var md opentsdb.MultiDataPoint
	c := 0
	if _, err := os.Stat(dscpath + "MetaConfig.mof"); os.IsNotExist(err) {
		c = 1
	}
	Add(&md, dscLCM+"configured", c, nil, metadata.Gauge, metadata.StatusCode, descDSCLCMConfigured)
	if c == 1 {
		return md, nil
	}
	for filename, filetype := range mapMofFiles {
		tags := opentsdb.TagSet{"type": filetype}
		s := int64(-1)
		l := int64(-1)
		if fi, fierr := os.Stat(dscpath + filename); fierr == nil {
			s = fi.Size()
			l = time.Now().Unix() - fi.ModTime().Unix()
		}
		Add(&md, dscMof+"size", s, tags, metadata.Gauge, metadata.Bytes, descDSCMofSize)
		Add(&md, dscMof+"last_modified", l, tags, metadata.Gauge, metadata.Second, descDSCMofModified)
	}
	return md, nil
}

const (
	descDSCLCMConfigured = "Indicates if DSC Local Configuration Manager is configured: 0=configured, 1=not configured. If the LCM is not configured then the rest of the dsc.* metrics will be skipped on that server."
	descDSCMofSize       = "Size of the mof file in bytes or -1 if file does not exist."
	descDSCMofModified   = "Number of seconds since the mof file was last modified or -1 if file does not exist."
)

func c_dsc_status() (opentsdb.MultiDataPoint, error) {
	var md opentsdb.MultiDataPoint
	if _, err := os.Stat(dscpath + "MetaConfig.mof"); os.IsNotExist(err) {
		continue
	}

	//We can't use the normal query syntax, as the default "intances" don't exist.
	//var dsc []MSFT_DSCConfigurationStatus
	//var q = wmi.CreateQuery(&dsc, "")
	//err := queryWmi(q, &dst)
	//if err != nil {
	//	return nil, err
	//}

	//Instead we need to modify "github.com/stackexchange/wmi" to allow us tocall a static method on the MSFT_DSCLocalConfigurationManager class
	//The MSFT_DSCLocalConfigurationManager.GetConfigurationStatus() method returns a Microsoft.Management.Infrastructure.CimMethodResult#MSFT_DSCLocalConfigurationManager#GetConfigurationStatus
	//object (at least in powershell it does), which has a ReturnValue property that is a MSFT_DSCConfigurationStatus instance.

	//Powershell that works on ny-rdp01
	//$dsclcm = Get-CimClass -Namespace ROOT\Microsoft\Windows\DesiredStateConfiguration MSFT_DSCLocalConfigurationManager
	//icim -CimClass $dsclcm -MethodName GetConfigurationStatus -ComputerName localhost | select -ExpandProperty ConfigurationStatus | fl *

}

const (
	descWinDSCDurationInSeconds          = "Time taken to process entire configuration."
	descWinDSCError                      = "Error encountered in local configuration manager during configuration."
	descWinDSCLCMVersion                 = "Version of LCM at time of configuration."
	descWinDSCMetaConfiguration          = "Meta-Configuration information at time of configuration."
	descWinDSCMetaData                   = "Meta data of configuration."
	descWinDSCMode                       = "Mode of configuration."
	descWinDSCNumberOfResources          = "Total number of resources in configuration."
	descWinDSCRebootRequested            = "Reboot was requested during configuration."
	descWinDSCResourcesInDesiredState    = "Resources successfully configured in the configuration."
	descWinDSCResourcesNotInDesiredState = "Resources failed in the configuration."
	descWinDSCStartDate                  = "Date and time when the configuration was started."
	descWinDSCStatus                     = "Status of configuration."
	descWinDSCType                       = "Type of Configuration."
)

type MSFT_DSCConfigurationStatus struct {
	DurationInSeconds uint32
	Error             string
	LCMVersion        string
	//MetaConfiguration object:MSFT_DSCMetaConfiguration
	MetaData          string
	Mode              string
	NumberOfResources uint32
	RebootRequested   boolean
	//ResourcesInDesiredState object:MSFT_ResourceInDesiredState
	//ResourcesNotInDesiredState object:MSFT_ResourceNotInDesiredState
	StartDate string
	Status    string
	Type      string
}
