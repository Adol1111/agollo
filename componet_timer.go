package agollo

import (
	"time"
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	t2 := time.NewTimer(refresh_interval)
	for {
		select {
		case <-t2.C:
			notifySyncConfigServices()
			t2.Reset(refresh_interval)
		}
	}
}

func SyncConfig() error {
	return autoSyncConfigServices()
}


func autoSyncConfigServicesSuccessCallBack(responseBody []byte)(o interface{},err error){
	apolloConfig,err:=createApolloConfigWithJson(responseBody)

	if err!=nil{
		logger.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}

	updateApolloConfig(apolloConfig)

	return nil,nil
}

func autoSyncConfigServices() error {
	appConfig := GetAppConfig()
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	for _, namespace := range appConfig.NamespaceNames {
		urlSuffix := getConfigUrlSuffix(appConfig, namespace)

		_, err := requestRecovery(appConfig, &ConnectConfig{
			Uri: urlSuffix,
		}, &CallBack{
			SuccessCallBack:   autoSyncConfigServicesSuccessCallBack,
			NotModifyCallBack: warpTouchApolloConfigCache(namespace),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func warpTouchApolloConfigCache(namespace string) func() error {
	return func() error {
		return touchApolloConfigCache(namespace)
	}
}
