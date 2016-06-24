package orm

import (
	"regexp"
	"strings"

	"github.com/lfq7413/tomato/errs"
	"github.com/lfq7413/tomato/storage"
	"github.com/lfq7413/tomato/types"
	"github.com/lfq7413/tomato/utils"
)

// clpValidKeys 类级别的权限 列表
var clpValidKeys = []string{"find", "get", "create", "update", "delete", "addField", "readUserFields", "writeUserFields"}

// SystemClasses 系统表
var SystemClasses = []string{"_User", "_Installation", "_Role", "_Session", "_Product", "_PushStatus"}

var volatileClasses = []string{"_PushStatus", "_Hooks", "_GlobalConfig"}

// DefaultColumns 所有类的默认字段，以及系统类的默认字段
var DefaultColumns = map[string]types.M{
	"_Default": types.M{
		"objectId":  types.M{"type": "String"},
		"createdAt": types.M{"type": "Date"},
		"updatedAt": types.M{"type": "Date"},
		"ACL":       types.M{"type": "ACL"},
	},
	"_User": types.M{
		"username":      types.M{"type": "String"},
		"password":      types.M{"type": "String"},
		"email":         types.M{"type": "String"},
		"emailVerified": types.M{"type": "Boolean"},
	},
	"_Installation": types.M{
		"installationId":   types.M{"type": "String"},
		"deviceToken":      types.M{"type": "String"},
		"channels":         types.M{"type": "Array"},
		"deviceType":       types.M{"type": "String"},
		"pushType":         types.M{"type": "String"},
		"GCMSenderId":      types.M{"type": "String"},
		"timeZone":         types.M{"type": "String"},
		"localeIdentifier": types.M{"type": "String"},
		"badge":            types.M{"type": "Number"},
		"appVersion":       types.M{"type": "String"},
		"appName":          types.M{"type": "String"},
		"appIdentifier":    types.M{"type": "String"},
		"parseVersion":     types.M{"type": "String"},
	},
	"_Role": types.M{
		"name":  types.M{"type": "String"},
		"users": types.M{"type": "Relation", "targetClass": "_User"},
		"roles": types.M{"type": "Relation", "targetClass": "_Role"},
	},
	"_Session": types.M{
		"restricted":     types.M{"type": "Boolean"},
		"user":           types.M{"type": "Pointer", "targetClass": "_User"},
		"installationId": types.M{"type": "String"},
		"sessionToken":   types.M{"type": "String"},
		"expiresAt":      types.M{"type": "Date"},
		"createdWith":    types.M{"type": "Object"},
	},
	"_Product": types.M{
		"productIdentifier": types.M{"type": "String"},
		"download":          types.M{"type": "File"},
		"downloadName":      types.M{"type": "String"},
		"icon":              types.M{"type": "File"},
		"order":             types.M{"type": "Number"},
		"title":             types.M{"type": "String"},
		"subtitle":          types.M{"type": "String"},
	},
	"_PushStatus": types.M{
		"pushTime":      types.M{"type": "String"},
		"source":        types.M{"type": "String"}, // rest or webui
		"query":         types.M{"type": "String"}, // the stringified JSON query
		"payload":       types.M{"type": "Object"}, // the JSON payload,
		"title":         types.M{"type": "String"},
		"expiry":        types.M{"type": "Number"},
		"status":        types.M{"type": "String"},
		"numSent":       types.M{"type": "Number"},
		"numFailed":     types.M{"type": "Number"},
		"pushHash":      types.M{"type": "String"},
		"errorMessage":  types.M{"type": "Object"},
		"sentPerType":   types.M{"type": "Object"},
		"failedPerType": types.M{"type": "Object"},
	},
}

// requiredColumns 类必须要有的字段
var requiredColumns = map[string][]string{
	"_Product": []string{"productIdentifier", "icon", "order", "title", "subtitle"},
	"_Role":    []string{"name", "ACL"},
}

// Schema schema 操作对象
type Schema struct {
	dbAdapter storage.Adapter
	data      types.M // data 保存类的字段信息，类型为 API 类型
	perms     types.M // perms 保存类的操作权限
}

// AddClassIfNotExists 添加类定义，包含默认的字段
func (s *Schema) AddClassIfNotExists(className string, fields types.M, classLevelPermissions types.M) (types.M, error) {
	err := s.validateNewClass(className, fields, classLevelPermissions)
	if err != nil {
		return nil, err
	}

	schema := types.M{
		"className":             className,
		"fields":                fields,
		"classLevelPermissions": classLevelPermissions,
	}
	result, err := s.dbAdapter.CreateClass(className, convertSchemaToAdapterSchema(schema))
	result = convertAdapterSchemaToParseSchema(result)
	if err != nil {
		if errs.GetErrorCode(err) == errs.DuplicateValue {
			return nil, errs.E(errs.InvalidClassName, "Class "+className+" already exists.")
		}
		return nil, err
	}

	return result, nil
}

// UpdateClass 更新类
func (s *Schema) UpdateClass(className string, submittedFields types.M, classLevelPermissions types.M) (types.M, error) {
	schema, err := s.GetOneSchema(className, false)
	if err != nil {
		return nil, err
	}
	if schema == nil || len(schema) == 0 || utils.M(schema["fields"]) == nil {
		return nil, errs.E(errs.InvalidClassName, "Class "+className+" does not exist.")
	}

	// 组装已存在的字段
	existingFields := utils.M(schema["fields"])

	// 校验对字段的操作是否合法
	for name, v := range submittedFields {
		field := utils.M(v)
		op := utils.S(field["__op"])
		if existingFields[name] != nil && op != "Delete" {
			// 字段已存在，不能更新
			return nil, errs.E(errs.ClassNotEmpty, "Field "+name+" exists, cannot update.")
		}
		if existingFields[name] == nil && op == "Delete" {
			// 字段不存在，不能删除
			return nil, errs.E(errs.ClassNotEmpty, "Field "+name+" does not exist, cannot delete.")
		}
	}

	delete(existingFields, "_rperm")
	delete(existingFields, "_wperm")

	// 组装写入数据库的数据
	newSchema := buildMergedSchemaObject(existingFields, submittedFields)
	existingFieldNames := []string{}
	for k := range existingFields {
		existingFieldNames = append(existingFieldNames, k)
	}
	err = s.validateSchemaData(className, newSchema, classLevelPermissions, existingFieldNames)
	if err != nil {
		return nil, err
	}

	// 删除指定字段，并统计需要插入的字段
	insertedFields := []string{}
	for name, v := range submittedFields {
		field := utils.M(v)
		op := utils.S(field["__op"])
		if op == "Delete" {
			err := s.deleteField(name, className)
			if err != nil {
				return nil, err
			}
		} else {
			insertedFields = append(insertedFields, name)
		}
	}

	// 重新加载修改过的数据
	s.reloadData()

	// 校验并插入字段
	for _, fieldName := range insertedFields {
		fieldType := submittedFields[fieldName].(map[string]interface{})
		err := s.enforceFieldExists(className, fieldName, fieldType, false)
		if err != nil {
			return nil, err
		}
	}

	// 设置 CLP
	err = s.setPermissions(className, classLevelPermissions, newSchema)
	if err != nil {
		return nil, err
	}

	return types.M{
		"className":             className,
		"fields":                s.data[className],
		"classLevelPermissions": s.perms[className],
	}, nil
}

// deleteField 从类定义中删除指定的字段，并删除对象中的数据
func (s *Schema) deleteField(fieldName string, className string) error {
	if ClassNameIsValid(className) == false {
		return errs.E(errs.InvalidClassName, InvalidClassNameMessage(className))
	}
	if fieldNameIsValid(fieldName) == false {
		return errs.E(errs.InvalidKeyName, "invalid field name: "+fieldName)
	}
	if fieldNameIsValidForClass(fieldName, className) == false {
		return errs.E(errs.ChangedImmutableFieldError, "field "+fieldName+" cannot be changed")
	}

	schema, err := s.GetOneSchema(className, false)
	if err != nil {
		return err
	}
	if schema == nil || len(schema) == 0 || utils.M(schema["fields"]) == nil {
		return errs.E(errs.InvalidClassName, "Class "+className+" does not exist.")
	}

	fields := utils.M(schema["fields"])
	if fields[fieldName] == nil {
		return errs.E(errs.ClassNotEmpty, "Field "+fieldName+" does not exist, cannot delete.")
	}

	// 根据字段属性进行相应 对象数据 删除操作
	if fieldType := utils.M(fields[fieldName]); fieldType != nil {
		if utils.S(fieldType["type"]) == "Relation" {
			// 删除表数据与 schema 中的对应字段
			err := s.dbAdapter.DeleteFields(className, schema, []string{fieldName})
			if err != nil {
				return err
			}
			// 删除 _Join table 数据
			_, err = s.dbAdapter.DeleteClass("_Join:" + fieldName + ":" + className)
			if err != nil {
				return err
			}
			return nil
		}
	}

	// 删除其他类型字段 对应的对象数据
	return s.dbAdapter.DeleteFields(className, schema, []string{fieldName})
}

// validateObject 校验对象是否合法
func (s *Schema) validateObject(className string, object, query types.M) error {
	geocount := 0
	err := s.EnforceClassExists(className)
	if err != nil {
		return err
	}

	for fieldName, v := range object {
		if v == nil {
			continue
		}
		expected, err := getType(v)
		if err != nil {
			return err
		}
		if expected == nil {
			continue
		}
		if expected["type"].(string) == "GeoPoint" {
			geocount++
		}
		if geocount > 1 {
			// 只能有一个 geopoint
			return errs.E(errs.IncorrectType, "there can only be one geopoint field in a class")
		}
		if fieldName == "ACL" {
			// 每个对象都隐含 ACL 字段
			continue
		}
		// 添加字段
		err = s.enforceFieldExists(className, fieldName, expected, false)
		if err != nil {
			return err
		}
	}

	err = thenValidateRequiredColumns(s, className, object, query)
	if err != nil {
		return err
	}
	return nil
}

// testBaseCLP 校验用户是否有权限对表进行指定操作
func (s *Schema) testBaseCLP(className string, aclGroup []string, operation string) bool {
	if s.perms[className] == nil && utils.M(s.perms[className])[operation] == nil {
		return true
	}
	classPerms := utils.M(s.perms[className])
	perms := utils.M(classPerms[operation])
	// 当前操作的权限是公开的
	if _, ok := perms["*"]; ok {
		return true
	}

	// 查找 acl 中的角色信息是否在权限列表中，找到一个即可
	found := false
	for _, v := range aclGroup {
		if _, ok := perms[v]; ok {
			found = true
			break
		}
	}
	if found {
		return true
	}

	return false
}

// validatePermission 校验对指定类的操作权限
func (s *Schema) validatePermission(className string, aclGroup []string, operation string) error {
	if s.testBaseCLP(className, aclGroup, operation) {
		return nil
	}

	if s.perms[className] == nil && utils.M(s.perms[className])[operation] == nil {
		return nil
	}
	classPerms := utils.M(s.perms[className])

	var permissionField string
	if operation == "get" || operation == "find" {
		permissionField = "readUserFields"
	} else {
		permissionField = "writeUserFields"
	}

	if permissionField == "writeUserFields" && operation == "create" {
		return errs.E(errs.OperationForbidden, "Permission denied for this action.")
	}

	if v, ok := classPerms[permissionField].([]interface{}); ok && len(v) > 0 {
		return nil
	}

	return errs.E(errs.OperationForbidden, "Permission denied for this action.")
}

// EnforceClassExists 校验类名
func (s *Schema) EnforceClassExists(className string) error {
	if s.data[className] != nil {
		return nil
	}

	// 添加不存在的类定义
	_, err := s.AddClassIfNotExists(className, types.M{}, types.M{})
	if err != nil {

	}
	s.reloadData()

	if s.data[className] != nil {
		return nil
	}

	return errs.E(errs.InvalidJSON, "Failed to add "+className)
}

// validateNewClass 校验新建的类
func (s *Schema) validateNewClass(className string, fields types.M, classLevelPermissions types.M) error {
	if s.data[className] != nil {
		return errs.E(errs.InvalidClassName, "Class "+className+" already exists.")
	}

	if ClassNameIsValid(className) == false {
		return errs.E(errs.InvalidClassName, InvalidClassNameMessage(className))
	}

	return s.validateSchemaData(className, fields, classLevelPermissions, []string{})
}

// validateSchemaData 校验 Schema 数据
func (s *Schema) validateSchemaData(className string, fields types.M, classLevelPermissions types.M, existingFieldNames []string) error {
	for fieldName, v := range fields {
		exist := false
		for _, k := range existingFieldNames {
			if fieldName == k {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		if fieldNameIsValid(fieldName) == false {
			return errs.E(errs.InvalidKeyName, "invalid field name: "+fieldName)
		}
		if fieldNameIsValidForClass(fieldName, className) == false {
			return errs.E(errs.ChangedImmutableFieldError, "field "+fieldName+" cannot be added")
		}
		err := fieldTypeIsInvalid(v.(map[string]interface{}))
		if err != nil {
			return err
		}
	}

	if DefaultColumns[className] != nil {
		for fieldName, v := range DefaultColumns[className] {
			fields[fieldName] = v
		}
	}

	geoPoints := []string{}
	for key, v := range fields {
		if v != nil {
			fieldData := v.(map[string]interface{})
			if fieldData["type"].(string) == "GeoPoint" {
				geoPoints = append(geoPoints, key)
			}
		}
	}
	if len(geoPoints) > 1 {
		return errs.E(errs.IncorrectType, "currently, only one GeoPoint field may exist in an object. Adding "+geoPoints[1]+" when "+geoPoints[0]+" already exists.")
	}

	return validateCLP(classLevelPermissions, fields)
}

// validateRequiredColumns 校验必须的字段
func (s *Schema) validateRequiredColumns(className string, object, query types.M) error {
	columns := requiredColumns[className]
	if columns == nil || len(columns) == 0 {
		return nil
	}

	missingColumns := []string{}
	for _, column := range columns {
		if query != nil && query["objectId"] != nil {
			// 类必须的字段，不能进行删除操作
			if object[column] != nil && utils.M(object[column]) != nil {
				o := utils.M(object[column])
				if utils.S(o["__op"]) == "Delete" {
					missingColumns = append(missingColumns, column)
				}
			}
			continue
		}
		// 不能缺少必须的字段
		if object[column] == nil {
			missingColumns = append(missingColumns, column)
		}
	}

	if len(missingColumns) > 0 {
		return errs.E(errs.IncorrectType, missingColumns[0]+" is required.")
	}
	return nil
}

// enforceFieldExists 校验并插入字段，freeze 为 true 时不进行修改
func (s *Schema) enforceFieldExists(className, fieldName string, fieldtype types.M, freeze bool) error {
	if strings.Index(fieldName, ".") > 0 {
		fieldName = strings.Split(fieldName, ".")[0]
		fieldtype = types.M{
			"type": "Object",
		}
	}

	if fieldNameIsValid(fieldName) == false {
		return errs.E(errs.InvalidKeyName, "Invalid field name: "+fieldName)
	}

	if fieldtype == nil || len(fieldtype) == 0 {
		return nil
	}

	s.reloadData()

	expectedType := s.getExpectedType(className, fieldName)
	if expectedType != nil {
		if dbTypeMatchesObjectType(expectedType, fieldtype) == false {
			return errs.E(errs.IncorrectType, "schema mismatch for "+className+"."+fieldName+"; expected "+utils.S(expectedType["type"])+" but got "+utils.S(fieldtype["type"]))
		}
	}

	err := s.dbAdapter.AddFieldIfNotExists(className, fieldName, fieldtype)
	if err != nil {
		// 失败时也需要重新加载数据，因为这时候可能有其他客户端更新了字段
		// s.reloadData()
		// return err
	}

	s.reloadData()
	// 再次尝试校验字段
	if dbTypeMatchesObjectType(s.getExpectedType(className, fieldName), fieldtype) == false {
		return errs.E(errs.InvalidJSON, "Could not add field "+fieldName)
	}
	return nil
}

// setPermissions 给指定类设置权限
func (s *Schema) setPermissions(className string, perms types.M, newSchema types.M) error {
	if perms == nil {
		return nil
	}
	err := validateCLP(perms, newSchema)
	if err != nil {
		return err
	}
	err = s.dbAdapter.SetClassLevelPermissions(className, perms)
	if err != nil {
		return err
	}
	s.reloadData()
	return nil
}

// HasClass Schema 中是否存在类定义
func (s *Schema) HasClass(className string) bool {
	s.reloadData()
	return s.data[className] != nil
}

// getExpectedType 获取期望的字段类型
func (s *Schema) getExpectedType(className, fieldName string) types.M {
	if s.data != nil && s.data[className] != nil {
		cls := utils.M(s.data[className])
		expectedType := utils.M(cls[fieldName])
		if utils.S(expectedType["type"]) == "map" {
			expectedType["type"] = "Object"
		}
		return expectedType
	}
	return nil
}

// reloadData 从数据库加载表信息
func (s *Schema) reloadData() {
	s.data = types.M{}
	s.perms = types.M{}
	allSchemas, err := s.GetAllClasses()
	if err != nil {
		return
	}
	for _, schema := range allSchemas {
		s.data[utils.S(schema["className"])] = injectDefaultSchema(schema)["fields"]
		s.perms[utils.S(schema["className"])] = schema["classLevelPermissions"]
	}

	for _, className := range volatileClasses {
		sch := types.M{
			"className":             className,
			"fields":                types.M{},
			"classLevelPermissions": types.M{},
		}
		s.data[className] = injectDefaultSchema(sch)
	}
}

// GetAllClasses ...
func (s *Schema) GetAllClasses() ([]types.M, error) {
	allSchemas, err := s.dbAdapter.GetAllClasses()
	if err != nil {
		return nil, err
	}
	schems := []types.M{}
	for _, v := range allSchemas {
		schems = append(schems, injectDefaultSchema(v))
	}
	return schems, nil
}

// GetOneSchema allowVolatileClasses 默认为 false
func (s *Schema) GetOneSchema(className string, allowVolatileClasses bool) (types.M, error) {
	if allowVolatileClasses {
		for _, name := range volatileClasses {
			if name == className {
				return s.data[className].(map[string]interface{}), nil
			}
		}
	}

	schema, err := s.dbAdapter.GetClass(className)
	if err != nil {
		return nil, err
	}
	return injectDefaultSchema(schema), nil
}

// thenValidateRequiredColumns 校验必须的字段
func thenValidateRequiredColumns(schema *Schema, className string, object, query types.M) error {
	return schema.validateRequiredColumns(className, object, query)
}

// getType 获取对象的格式
func getType(obj interface{}) (types.M, error) {
	switch obj.(type) {
	case bool:
		return types.M{"type": "Boolean"}, nil
	case string:
		return types.M{"type": "String"}, nil
	case float64:
		return types.M{"type": "Number"}, nil
	case map[string]interface{}, []interface{}:
		return getObjectType(obj)
	default:
		return nil, errs.E(errs.IncorrectType, "bad obj. can not get type")
	}
}

// getObjectType 获取对象格式 仅处理 slice 与 map
func getObjectType(obj interface{}) (types.M, error) {
	if utils.A(obj) != nil {
		return types.M{"type": "Array"}, nil
	}
	if utils.M(obj) != nil {
		object := utils.M(obj)
		if object["__type"] != nil {
			t := utils.S(object["__type"])
			switch t {
			case "Pointer":
				if object["className"] != nil {
					return types.M{
						"type":        "Pointer",
						"targetClass": object["className"],
					}, nil
				}
			case "File":
				if object["name"] != nil {
					return types.M{"type": "File"}, nil
				}
			case "Date":
				if object["iso"] != nil {
					return types.M{"type": "Date"}, nil
				}
			case "GeoPoint":
				if object["latitude"] != nil && object["longitude"] != nil {
					return types.M{"type": "Geopoint"}, nil
				}
			case "Bytes":
				if object["base64"] != nil {
					return types.M{"type": "Bytes"}, nil
				}
			default:
				// 无效的类型
				return nil, errs.E(errs.IncorrectType, "This is not a valid "+t)
			}
		}
		if object["$ne"] != nil {
			return getObjectType(object["$ne"])
		}
		if object["__op"] != nil {
			op := utils.S(object["__op"])
			switch op {
			case "Increment":
				return types.M{"type": "Number"}, nil
			case "Delete":
				return nil, nil
			case "Add", "AddUnique", "Remove":
				return types.M{"type": "Array"}, nil
			case "AddRelation", "RemoveRelation":
				objects := utils.A(object["objects"])
				o := utils.M(objects[0])
				return types.M{
					"type":        "Relation",
					"targetClass": utils.S(o["className"]),
				}, nil
			case "Batch":
				ops := utils.A(object["ops"])
				return getObjectType(ops[0])
			default:
				// 无效操作
				return nil, errs.E(errs.IncorrectType, "unexpected op: "+op)
			}
		}
	}

	return types.M{"type": "object"}, nil
}

// ClassNameIsValid 校验类名，可以是系统内置类、join 类
// 数字字母组合，以及下划线，但不能以下划线或字母开头
func ClassNameIsValid(className string) bool {
	for _, k := range SystemClasses {
		if className == k {
			return true
		}
	}
	return joinClassIsValid(className) ||
		fieldNameIsValid(className) // 类名与字段名的规则相同
}

// InvalidClassNameMessage ...
func InvalidClassNameMessage(className string) string {
	return "Invalid classname: " + className + ", classnames can only have alphanumeric characters and _, and must start with an alpha character "
}

var joinClassRegex = `^_Join:[A-Za-z0-9_]+:[A-Za-z0-9_]+`

// joinClassIsValid 校验 join 表名， _Join:abc:abc
func joinClassIsValid(className string) bool {
	b, _ := regexp.MatchString(joinClassRegex, className)
	return b
}

var classAndFieldRegex = `^[A-Za-z][A-Za-z0-9_]*$`

// fieldNameIsValid 校验字段名或者类名，数字字母下划线，不以数字下划线开头
func fieldNameIsValid(fieldName string) bool {
	b, _ := regexp.MatchString(classAndFieldRegex, fieldName)
	return b
}

// fieldNameIsValidForClass 校验能否添加指定字段到类中
func fieldNameIsValidForClass(fieldName string, className string) bool {
	// 字段名不合法不能添加
	if fieldNameIsValid(fieldName) == false {
		return false
	}
	// 默认字段不能添加
	if DefaultColumns["_Default"][fieldName] != nil {
		return false
	}
	// 当前类的默认字段不能添加
	if DefaultColumns[className] != nil && DefaultColumns[className][fieldName] != nil {
		return false
	}

	return true
}

var validNonRelationOrPointerTypes = []string{
	"Number",
	"String",
	"Boolean",
	"Date",
	"Object",
	"Array",
	"GeoPoint",
	"File",
}

// fieldTypeIsInvalid 检测字段类型是否合法
func fieldTypeIsInvalid(t types.M) error {
	var invalidJSONError = errs.E(errs.InvalidJSON, "invalid JSON")
	fieldType := ""
	if v, ok := t["type"].(string); ok {
		fieldType = v
	} else {
		return invalidJSONError
	}
	targetClass := ""
	if fieldType == "Pointer" || fieldType == "Relation" {
		if _, ok := t["targetClass"]; ok == false {
			return errs.E(errs.MissingRequiredFieldError, "type "+fieldType+" needs a class name")
		}
		if v, ok := t["targetClass"].(string); ok {
			targetClass = v
		} else {
			return invalidJSONError
		}
		if ClassNameIsValid(targetClass) == false {
			return errs.E(errs.InvalidClassName, InvalidClassNameMessage(targetClass))
		}
		return nil
	}

	in := false
	for _, v := range validNonRelationOrPointerTypes {
		if fieldType == v {
			in = true
			break
		}
	}
	if in == false {
		return errs.E(errs.IncorrectType, "invalid field type: "+fieldType)
	}

	return nil
}

// validateCLP 校验类级别权限
// 正常的 perms 格式如下
// {
// 	"get":{
// 		"user24id":true,
// 		"role:xxx":true,
// 		"*":true,
// 	},
// 	"delete":{...},
//  "readUserFields":{"aaa","bbb"}
// 	...
// }
func validateCLP(perms types.M, fields types.M) error {
	if perms == nil {
		return nil
	}

	for operation, perm := range perms {
		// 校验是否是系统规定的几种操作
		t := false
		for _, key := range clpValidKeys {
			if operation == key {
				t = true
				break
			}
		}
		if t == false {
			return errs.E(errs.InvalidJSON, operation+" is not a valid operation for class level permissions")
		}

		if operation == "readUserFields" || operation == "writeUserFields" {
			if p, ok := perm.([]interface{}); ok {
				for _, v := range p {
					key := v.(string)
					// 字段类型必须为指向 _User 的指针类型
					if fields[key] != nil {
						if t, ok := fields[key].(map[string]interface{}); ok {
							if t["type"].(string) == "Pointer" && t["targetClass"].(string) == "_User" {
								continue
							}
						}
					}
					return errs.E(errs.InvalidJSON, key+" is not a valid column for class level pointer permissions "+operation)
				}
				return nil
			}
			return errs.E(errs.InvalidJSON, "this perms[operation] is not a valid value for class level permissions "+operation)
		}

		for key, p := range utils.M(perm) {
			err := verifyPermissionKey(key)
			if err != nil {
				return err
			}
			if v, ok := p.(bool); ok {
				if v == false {
					return errs.E(errs.InvalidJSON, "false is not a valid value for class level permissions "+operation+":"+key+":false")
				}
			} else {
				return errs.E(errs.InvalidJSON, "this perm is not a valid value for class level permissions "+operation+":"+key+":perm")
			}
		}
	}
	return nil
}

// 24 alpha numberic chars + uppercase
var userIDRegex = `^[a-zA-Z0-9]{24}$`

// Anything that start with role
var roleRegex = `^role:.*`

// * permission
var publicRegex = `^\*$`

var permissionKeyRegex = []string{userIDRegex, roleRegex, publicRegex}

// verifyPermissionKey 校验 CLP 中各种操作包含的角色名是否合法
// 可以是24位的用户 ID，可以是角色名 role:abc ,可以是公共权限 *
func verifyPermissionKey(key string) error {
	for _, v := range permissionKeyRegex {
		if b, _ := regexp.MatchString(v, key); b {
			return nil
		}
	}
	return errs.E(errs.InvalidJSON, key+" is not a valid key for class level permissions")
}

// buildMergedSchemaObject 组装数据库类型的 existingFields 与 API 类型的 putRequest，
// 返回值中不包含默认字段，返回的是 API 类型的数据
func buildMergedSchemaObject(existingFields types.M, putRequest types.M) types.M {
	newSchema := types.M{}

	sysSchemaField := []string{}
	id := utils.S(existingFields["_id"])
	for k, v := range DefaultColumns {
		// 如果是系统预定义的表，则取出默认字段
		if k == id {
			for key := range v {
				sysSchemaField = append(sysSchemaField, key)
			}
			break
		}
	}

	// 处理已经存在的字段
	for oldField, v := range existingFields {
		// 仅处理以下五种字段以外的字段
		if oldField != "_id" &&
			oldField != "ACL" &&
			oldField != "updatedAt" &&
			oldField != "createdAt" &&
			oldField != "objectId" {
			// 不处理系统默认字段
			if len(sysSchemaField) > 0 {
				t := false
				for _, s := range sysSchemaField {
					if s == oldField {
						t = true
						break
					}
				}
				if t == true {
					continue
				}
			}
			// 处理要删除的字段，要删除的字段不加入返回数据中
			fieldIsDeleted := false
			if putRequest[oldField] != nil {
				op := utils.M(putRequest[oldField])
				if utils.S(op["__op"]) == "Delete" {
					fieldIsDeleted = true
				}
			}
			if fieldIsDeleted == false {
				newSchema[oldField] = v
			}
		}
	}

	// 处理需要更新的字段
	for newField, v := range putRequest {
		op := utils.M(v)
		// 不处理 objectId，不处理要删除的字段，跳过系统默认字段，其余字段加入返回数据中
		if newField != "objectId" && utils.S(op["__op"]) != "Delete" {
			if len(sysSchemaField) > 0 {
				t := false
				for _, s := range sysSchemaField {
					if s == newField {
						t = true
						break
					}
				}
				if t == true {
					continue
				}
			}
			newSchema[newField] = v
		}
	}

	return newSchema
}

// injectDefaultSchema 为 schema 添加默认字段
func injectDefaultSchema(schema types.M) types.M {
	if schema == nil {
		return nil
	}
	newSchema := types.M{}
	newfields := types.M{}
	fields := utils.M(schema["fields"])
	if fields == nil {
		fields = types.M{}
	}
	for k, v := range fields {
		newfields[k] = v
	}
	defaultFieldsSchema := DefaultColumns["_Default"]
	for k, v := range defaultFieldsSchema {
		newfields[k] = v
	}
	defaultSchema := DefaultColumns[utils.S(schema["className"])]
	if defaultSchema != nil {
		for k, v := range defaultSchema {
			newfields[k] = v
		}
	}
	newSchema["fields"] = newfields
	newSchema["className"] = schema["className"]
	newSchema["classLevelPermissions"] = schema["classLevelPermissions"]

	return newSchema
}

// convertSchemaToAdapterSchema 转换 schema 为 Adapter 使用的类型：添加默认字段，删除不必要的字段
// {
// 	ACL:{type:ACL}
// 	password:{type:string}
// 	key:{type:string}
// }
// ==>
// {
// 	key:{type:string}
// 	_rperm:{type:Array}
// 	_wperm:{type:Array}
// 	_hashed_password:{type:string}
// }
func convertSchemaToAdapterSchema(schema types.M) types.M {
	if schema == nil {
		return schema
	}
	schema = injectDefaultSchema(schema)
	if fields := utils.M(schema["fields"]); fields != nil {
		delete(fields, "ACL")
		fields["_rperm"] = types.M{"type": "Array"}
		fields["_wperm"] = types.M{"type": "Array"}
		if utils.S(schema["className"]) == "_User" {
			delete(fields, "password")
			fields["_hashed_password"] = types.M{"type": "String"}
		}
	}

	return schema
}

// convertAdapterSchemaToParseSchema 转换 Adapter 中使用的 schema 为普通类型
// {
// 	key:{type:string}
// 	_rperm:{type:Array}
// 	_wperm:{type:Array}
// 	_hashed_password:{type:string}
// }
// ==>
// {
// 	ACL:{type:ACL}
// 	password:{type:string}
// 	key:{type:string}
// }
func convertAdapterSchemaToParseSchema(schema types.M) types.M {
	if schema == nil {
		return schema
	}
	if fields := utils.M(schema["fields"]); fields != nil {
		delete(fields, "_rperm")
		delete(fields, "_wperm")
		fields["ACL"] = types.M{"type": "ACL"}
		if utils.S(schema["className"]) == "_User" {
			delete(fields, "authData")
			delete(fields, "_hashed_password")
			fields["password"] = types.M{"type": "String"}
		}
	}

	return schema
}

func dbTypeMatchesObjectType(dbType, objectType types.M) bool {
	if dbType == nil && objectType == nil {
		return true
	}
	if dbType != nil && objectType == nil {
		return false
	}
	if dbType == nil && objectType != nil {
		return false
	}
	if utils.S(dbType["type"]) != utils.S(objectType["type"]) {
		return false
	}
	if utils.S(dbType["targetClass"]) != utils.S(objectType["targetClass"]) {
		return false
	}
	if utils.S(dbType["type"]) == utils.S(objectType["type"]) {
		return true
	}
	return false
}

// Load 返回一个新的 Schema 结构体
func Load(adapter storage.Adapter) *Schema {
	schema := &Schema{
		dbAdapter: adapter,
	}
	schema.reloadData()
	return schema
}
