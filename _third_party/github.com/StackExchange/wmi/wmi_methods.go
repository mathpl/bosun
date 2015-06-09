package wmi

import (
	"reflect"
	"runtime"

	"bosun.org/_third_party/github.com/go-ole/go-ole"
	"bosun.org/_third_party/github.com/go-ole/go-ole/oleutil"
)

func (c *Client) InvokeMethodNamespace(query string, dst interface{}, connectServerArgs ...interface{}) error {
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return ErrInvalidEntityType
	}
	dv = dv.Elem()
	mat, elemType := checkMultiArg(dv)
	if mat == multiArgTypeInvalid {
		return ErrInvalidEntityType
	}

	lock.Lock()
	defer lock.Unlock()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		oleerr := err.(*ole.OleError)
		// S_FALSE           = 0x00000001 // CoInitializeEx was already called on this thread
		if oleerr.Code() != ole.S_OK && oleerr.Code() != 0x00000001 {
			return err
		}
	} else {
		// Only invoke CoUninitialize if the thread was not initizlied before.
		// This will allow other go packages based on go-ole play along
		// with this library.
		defer ole.CoUninitialize()
	}

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return err
	}
	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wmi.Release()

	// service is a SWbemServices
	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", connectServerArgs...)
	if err != nil {
		return err
	}
	service := serviceRaw.ToIDispatch()
	defer serviceRaw.Clear()

	//See WMI scripting at https://msdn.microsoft.com/en-us/library/aa384833(v=vs.85).aspx

	//Get MSFT_DSCLocalConfigurationManager object:  https://msdn.microsoft.com/en-us/library/aa393868(v=vs.85).aspx
	resultRaw, err := oleutil.CallMethod(service, "Get", "MSFT_DSCLocalConfigurationManager")
	if err != nil {
		return err
	}
	result := resultRaw.ToIDispatch()
	defer resultRaw.Clear()

	resultRaw2, err := oleutil.CallMethod(resultRaw, "GetConfigurationStatus")
	if err != nil {
		return err
	}
	result := resultRaw2.ToIDispatch()
	defer resultRaw2.Clear()

	//Figure out what ^ is and how to get a MSFT_DSCConfigurationStatus out of the ConfigurationStatus property.
	//We also probably want the MSFT_DSCMetaConfiguration in the MetaConfiguration property of MSFT_DSCConfigurationStatus
	// and at least the lenghts of the ResourcesInDesiredState/ResourcesNotInDesiredState

	// result is a SWBemObjectSet
	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", query)
	if err != nil {
		return err
	}
	result := resultRaw.ToIDispatch()
	defer resultRaw.Clear()

	count, err := oleInt64(result, "Count")
	if err != nil {
		return err
	}

	// Initialize a slice with Count capacity
	dv.Set(reflect.MakeSlice(dv.Type(), 0, int(count)))

	var errFieldMismatch error
	for i := int64(0); i < count; i++ {
		err := func() error {
			// item is a SWbemObject, but really a Win32_Process
			itemRaw, err := oleutil.CallMethod(result, "ItemIndex", i)
			if err != nil {
				return err
			}
			item := itemRaw.ToIDispatch()
			defer itemRaw.Clear()

			ev := reflect.New(elemType)
			if err = c.loadEntity(ev.Interface(), item); err != nil {
				if _, ok := err.(*ErrFieldMismatch); ok {
					// We continue loading entities even in the face of field mismatch errors.
					// If we encounter any other error, that other error is returned. Otherwise,
					// an ErrFieldMismatch is returned.
					errFieldMismatch = err
				} else {
					return err
				}
			}
			if mat != multiArgTypeStructPtr {
				ev = ev.Elem()
			}
			dv.Set(reflect.Append(dv, ev))
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return errFieldMismatch
}
