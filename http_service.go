package gglmm

import (
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	// ActionGetByID 根据ID拉取单个
	ActionGetByID = "GetByID"
	// ActionFirst 根据条件拉取单个
	ActionFirst = "First"
	// ActionList 列表
	ActionList = "List"
	// ActionPage 分页
	ActionPage = "page"
	// ActionCreate 保存
	ActionCreate = "create"
	// ActionUpdate 更新整体
	ActionUpdate = "Update"
	// ActionUpdateFields 更新多个字段
	ActionUpdateFields = "UpdateFields"
	// ActionRemove 软删除
	ActionRemove = "Remove"
	// ActionRestore 恢复
	ActionRestore = "Resotre"
	// ActionDestory 硬删除
	ActionDestory = "Destory"
)

// IDRegexp ID正则表达式
const IDRegexp = "{id:[0-9]+}"

var (
	// ReadActions 读操作
	ReadActions = []string{ActionGetByID, ActionFirst, ActionList, ActionPage}
	// WriteActions 写操作
	WriteActions = []string{ActionCreate, ActionUpdate, ActionUpdateFields}
	// DeleteActions 删除操作
	DeleteActions = []string{ActionRemove, ActionRestore, ActionDestory}
	// AdminActions 管理操作
	AdminActions = []string{ActionPage, ActionCreate, ActionUpdate, ActionRemove, ActionRestore}
	// AllActions 所有Action
	AllActions = []string{ActionGetByID, ActionFirst, ActionList, ActionPage, ActionCreate, ActionUpdate, ActionUpdateFields, ActionRemove, ActionRestore, ActionDestory}
)

// FilterFunc 过滤函数
type FilterFunc func(filters []Filter, r *http.Request) []Filter

// BeforeCreateFunc 保存前调用
type BeforeCreateFunc func(model interface{}) (interface{}, error)

// BeforeUpdateFunc 更新前调用
type BeforeUpdateFunc func(model interface{}, id int64) (interface{}, int64, error)

// BeforeDeleteFunc 删除前调用
type BeforeDeleteFunc func(model interface{}) (interface{}, error)

// HTTPService HTTP服务
type HTTPService struct {
	modelType        reflect.Type
	modelValue       reflect.Value
	filterFunc       FilterFunc
	beforeCreateFunc BeforeCreateFunc
	beforeUpdateFunc BeforeUpdateFunc
	beforeDeleteFunc BeforeDeleteFunc
}

// NewHTTPService 新建HTTP服务
func NewHTTPService(model interface{}) *HTTPService {
	if gormRepository == nil {
		log.Fatal(ErrGormRepositoryNotRegister)
	}
	return &HTTPService{
		modelType:  reflect.TypeOf(model),
		modelValue: reflect.ValueOf(model),
	}
}

// HandleFilterFunc 设置过滤参数函数
func (service *HTTPService) HandleFilterFunc(handler FilterFunc) *HTTPService {
	service.filterFunc = handler
	return service
}

// HandleBeforeCreateFunc 设置保存前执行函数
func (service *HTTPService) HandleBeforeCreateFunc(handler BeforeCreateFunc) *HTTPService {
	service.beforeCreateFunc = handler
	return service
}

// HandleBeforeUpdateFunc 设置更新前执行函数
func (service *HTTPService) HandleBeforeUpdateFunc(handler BeforeUpdateFunc) *HTTPService {
	service.beforeUpdateFunc = handler
	return service
}

// HandleBeforeDeleteFunc 设置更新前执行函数
func (service *HTTPService) HandleBeforeDeleteFunc(handler BeforeDeleteFunc) *HTTPService {
	service.beforeDeleteFunc = handler
	return service
}

// Action --
func (service *HTTPService) Action(action string) (*HTTPAction, error) {
	var path string
	var handlerFunc http.HandlerFunc
	var method string
	switch action {
	case ActionGetByID:
		path = "/" + IDRegexp
		handlerFunc = service.GetByID
		method = "GET"
	case ActionFirst:
		path = "/first"
		handlerFunc = service.First
		method = "POST"
	case ActionList:
		path = "/list"
		handlerFunc = service.List
		method = "POST"
	case ActionPage:
		path = "/page"
		handlerFunc = service.Page
		method = "POST"
	case ActionCreate:
		handlerFunc = service.Store
		method = "POST"
	case ActionUpdate:
		path = "/" + IDRegexp
		handlerFunc = service.Update
		method = "PUT"
	case ActionUpdateFields:
		path = "/" + IDRegexp
		handlerFunc = service.UpdateFields
		method = "PATCH"
	case ActionRemove:
		path = "/" + IDRegexp + "/remove"
		handlerFunc = service.Remove
		method = "DELETE"
	case ActionRestore:
		path = "/" + IDRegexp + "/restore"
		handlerFunc = service.Restore
		method = "DELETE"
	case ActionDestory:
		path = "/" + IDRegexp + "/destroy"
		handlerFunc = service.Destory
		method = "DELETE"
	}
	if handlerFunc != nil {
		return NewHTTPAction(path, handlerFunc, method), nil
	}
	return nil, ErrAction
}

// GetByID 单个
func (service *HTTPService) GetByID(w http.ResponseWriter, r *http.Request) {
	idRequest, err := DecodeIDRequest(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	model := reflect.New(service.modelType).Interface()
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(idRequest.ID, 10)
			if len(idRequest.Preloads) > 0 {
				cacheKey = cacheKey + ":" + strings.Join(idRequest.Preloads, "-")
			}
			if err := cacher.GetObj(cacheKey, model); err == nil {
				OkResponse().
					AddData(SingleKey(service.modelValue), model).
					JSON(w)
				return
			}
		}
	}
	if err = gormRepository.Get(model, idRequest); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(idRequest.ID, 10)
			if len(idRequest.Preloads) > 0 {
				cacheKey = cacheKey + ":" + strings.Join(idRequest.Preloads, "-")
			}
			cacher.Set(cacheKey, model)
		}
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// First 单个
func (service *HTTPService) First(w http.ResponseWriter, r *http.Request) {
	filterRequest, err := DecodeFilterRequest(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if service.filterFunc != nil {
		filterRequest.Filters = service.filterFunc(filterRequest.Filters, r)
	}
	model := reflect.New(service.modelType).Interface()
	if err = gormRepository.Get(model, filterRequest); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// List 列表
func (service *HTTPService) List(w http.ResponseWriter, r *http.Request) {
	filterRequest, err := DecodeFilterRequest(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if service.filterFunc != nil {
		filterRequest.Filters = service.filterFunc(filterRequest.Filters, r)
	}
	list := reflect.New(reflect.SliceOf(service.modelType)).Interface()
	if err = gormRepository.List(list, filterRequest); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	OkResponse().
		AddData(MultiKey(service.modelValue), list).
		JSON(w)
}

// Page 分页
func (service *HTTPService) Page(w http.ResponseWriter, r *http.Request) {
	pageRequest, err := DecodePageRequest(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if service.filterFunc != nil {
		pageRequest.Filters = service.filterFunc(pageRequest.Filters, r)
	}
	pageResponse := PageResponse{}
	pageResponse.List = reflect.New(reflect.SliceOf(service.modelType)).Interface()
	if err = gormRepository.Page(&pageResponse, pageRequest); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	OkResponse().
		AddData(MultiKey(service.modelValue), pageResponse.List).
		AddData("pagination", pageResponse.Pagination).
		JSON(w)
}

// Store 保存
func (service *HTTPService) Store(w http.ResponseWriter, r *http.Request) {
	model, err := DecodeModelPtr(r, service.modelType)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if service.beforeCreateFunc != nil {
		model, err = service.beforeCreateFunc(model)
		if err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
	}
	if err = gormRepository.Store(model); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// Update 更新整体
func (service *HTTPService) Update(w http.ResponseWriter, r *http.Request) {
	id, err := PathVarID(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	model, err := DecodeModelPtr(r, service.modelType)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if service.beforeUpdateFunc != nil {
		model, id, err = service.beforeUpdateFunc(model, id)
		if err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
	}
	if err = gormRepository.Update(model, id); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(id, 10)
			cacher.DelPattern(cacheKey)
		}
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// UpdateFields 更新整体
func (service *HTTPService) UpdateFields(w http.ResponseWriter, r *http.Request) {
	id, err := PathVarID(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	model := reflect.New(service.modelType).Interface()
	if err := gormRepository.Get(model, id); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	fields, err := DecodeModelPtr(r, service.modelType)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if service.beforeUpdateFunc != nil {
		fields, id, err = service.beforeUpdateFunc(fields, id)
		if err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
	}
	if err = gormRepository.UpdateFields(model, fields); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(id, 10)
			cacher.DelPattern(cacheKey)
		}
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// Remove 软删除
func (service *HTTPService) Remove(w http.ResponseWriter, r *http.Request) {
	id, err := PathVarID(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	model := reflect.New(service.modelType).Interface()
	if service.beforeDeleteFunc != nil {
		if err := gormRepository.Get(model, id); err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
		if _, err := service.beforeDeleteFunc(model); err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
	}
	if err = gormRepository.Remove(model, id); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(id, 10)
			cacher.DelPattern(cacheKey)
		}
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// Restore 恢复
func (service *HTTPService) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := PathVarID(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	model := reflect.New(service.modelType).Interface()
	if err = gormRepository.Restore(model, id); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(id, 10)
			cacher.DelPattern(cacheKey)
		}
	}
	OkResponse().
		AddData(SingleKey(service.modelValue), model).
		JSON(w)
}

// Destory 直接删除
func (service *HTTPService) Destory(w http.ResponseWriter, r *http.Request) {
	id, err := PathVarID(r)
	if err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	model := reflect.New(service.modelType).Interface()
	if service.beforeDeleteFunc != nil {
		if err := gormRepository.Get(model, id); err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
		if _, err := service.beforeDeleteFunc(model); err != nil {
			ErrorResponse(err.Error()).JSON(w)
			return
		}
	}
	if err = gormRepository.Destroy(model, id); err != nil {
		ErrorResponse(err.Error()).JSON(w)
		return
	}
	if cacher != nil {
		if SupportCache(service.modelValue) {
			cacheKey := service.modelType.Name() + ":" + strconv.FormatInt(id, 10)
			cacher.DelPattern(cacheKey)
		}
	}
	OkResponse().JSON(w)
}
