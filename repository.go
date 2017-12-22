package agollo

import (
	"strconv"
	"github.com/coocood/freecache"
	"sync"
)

const (
	empty  = ""

	//50m
	apolloConfigCacheSize=50*1024*1024

	//1 minute
	configCacheExpireTime=120
)
var (
	apolloConfigConnMap sync.Map
	//currentConnApolloConfig = &ApolloConnConfig{}

	//config from apollo
	apolloConfigCacheMap sync.Map
	cacheLock sync.Mutex
	//apolloConfigCache = freecache.NewCache(apolloConfigCacheSize)
)

func updateApolloConfig(apolloConfig *ApolloConfig) {
	if apolloConfig==nil{
		logger.Error("apolloConfig is null,can't update!")
		return
	}

	apolloConfigConnMap.Store(apolloConfig.NamespaceName, &apolloConfig.ApolloConnConfig)

	//get change list
	changeList := updateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime,apolloConfig.NamespaceName)

	if len(changeList) > 0 {
		//create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName)

		//push change event to channel
		pushChangeEvent(event)
	}

	//update apollo connection config
}

func updateApolloConfigCache(configurations map[string]string,expireTime int,namespace string ) map[string]*ConfigChange {
	apolloConfigCache:=GetApolloConfigCache(namespace)

	if (configurations==nil||len(configurations)==0)&& apolloConfigCache.EntryCount() == 0 {
		return nil
	}

	//get old keys
	mp := map[string]bool{}
	it := apolloConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		mp[string(en.Key)] = true
	}

	changes:=make(map[string]*ConfigChange)

	if configurations != nil {
		// update new
		// keys
		for key, value := range configurations {
			//key state insert or update
			//insert
			if !mp[key] {
				changes[key] = createAddConfigChange(value)
			} else {
				//update
				oldValue, _ := apolloConfigCache.Get([]byte(key))
				if string(oldValue) != value {
					changes[key] = createModifyConfigChange(string(oldValue), value)
				}
			}

			apolloConfigCache.Set([]byte(key), []byte(value), expireTime)
			delete(mp, string(key))
		}
	}

	// remove del keys
	for key := range mp {
		//get old value and del
		oldValue, _ := apolloConfigCache.Get([]byte(key))
		changes[key]=createDeletedConfigChange(string(oldValue))

		apolloConfigCache.Del([]byte(key))
	}

	return changes
}

//base on changeList create Change event
func createConfigChangeEvent(changes map[string]*ConfigChange,nameSpace string) *ChangeEvent {
	return &ChangeEvent{
		Namespace:nameSpace,
		Changes:changes,
	}
}

func touchApolloConfigCache(namespace string) error{
	updateApolloConfigCacheTime(configCacheExpireTime,namespace)
	return nil
}

func updateApolloConfigCacheTime(expireTime int, namespace string)  {
	apolloConfigCache:=GetApolloConfigCache(namespace)
	it := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it.Next()
		if entry==nil{
			break
		}
		apolloConfigCache.Set([]byte(entry.Key),[]byte(entry.Value),expireTime)
	}
}

func GetApolloConfigCache(namespace string) *freecache.Cache {
	v,ok:=apolloConfigCacheMap.Load(namespace)
	if !ok {
		cacheLock.Lock()
		defer cacheLock.Unlock()

		v,ok =apolloConfigCacheMap.Load(namespace)
		if !ok {
			cache := freecache.NewCache(apolloConfigCacheSize)
			apolloConfigCacheMap.Store(namespace, cache)
			v,ok =apolloConfigCacheMap.Load(namespace)
		}
	}

	cache, _ := v.(*freecache.Cache)
	return cache
}

func GetCurrentApolloConfig(namespace string) *ApolloConnConfig {
	v, ok := apolloConfigConnMap.Load(namespace)

	if !ok {
		return nil
	}

	config, _ := v.(*ApolloConnConfig)
	return config
}

func getConfigValue(key string, namespace string) interface{}  {
	apolloConfigCache:= GetApolloConfigCache(namespace)
	value,err:=apolloConfigCache.Get([]byte(key))
	if err!=nil{
		logger.Error("get config value fail!err:",err)
		return empty
	}

	return string(value)
}


func getValue(key string, namespace string)string{
	value:=getConfigValue(key, namespace)
	if value==nil{
		return empty
	}

	return value.(string)
}

func GetStringValue(key string,defaultValue string)string{
	value:=getValue(key, default_namespace)
	if value==empty{
		return defaultValue
	}

	return value
}

func GetIntValue(key string,defaultValue int) int {
	value :=getValue(key, default_namespace)

	i,err:=strconv.Atoi(value)
	if err!=nil{
		logger.Debug("convert to int fail!error:",err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string,defaultValue float64) float64 {
	value :=getValue(key,default_namespace)

	i,err:=strconv.ParseFloat(value,64)
	if err!=nil{
		logger.Debug("convert to float fail!error:",err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string,defaultValue bool) bool {
	value :=getValue(key,default_namespace )

	b,err:=strconv.ParseBool(value)
	if err!=nil{
		logger.Debug("convert to bool fail!error:",err)
		return defaultValue
	}

	return b
}

func GetStringValueWithNameSpace(key string,defaultValue string, namespace string)string{
	value:=getValue(key, namespace)
	if value==empty{
		return defaultValue
	}

	return value
}

func GetIntValueWithNameSpace(key string,defaultValue int, namespace string) int {
	value :=getValue(key, namespace)

	i,err:=strconv.Atoi(value)
	if err!=nil{
		logger.Debug("convert to int fail!error:",err)
		return defaultValue
	}

	return i
}

func GetFloatValueWithNameSpace(key string,defaultValue float64, namespace string) float64 {
	value :=getValue(key,namespace)

	i,err:=strconv.ParseFloat(value,64)
	if err!=nil{
		logger.Debug("convert to float fail!error:",err)
		return defaultValue
	}

	return i
}

func GetBoolValueWithNameSpace(key string,defaultValue bool, namespace string) bool {
	value :=getValue(key,namespace)

	b,err:=strconv.ParseBool(value)
	if err!=nil{
		logger.Debug("convert to bool fail!error:",err)
		return defaultValue
	}

	return b
}