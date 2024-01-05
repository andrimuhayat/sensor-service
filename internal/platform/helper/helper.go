package helper

import (
	"dario.cat/mergo"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	DDMMYYYYhhmmss  = "2006-01-02 15:04:05"
	Secretkey       = "secretkeyjwt"
	oneToMany       = "ONETOMANY"
	oneToOne        = "ONETOONE"
	manyToOne       = "MANYTOONE"
	protectedFields = []string{"created_at", "updated_at", "started_at", "ended_at", "started_at", "ended_at", "new_user_started_at", "new_user_ended_at", "timestamp", "id"}
)

func StringBoolToBool(value string) bool {
	if value == "true" {
		return true
	}

	return false
}

func LimitOffset(limit string, offset string) (int, int) {
	if limit == "" {
		limit = "10"
	}
	lmt, _ := strconv.Atoi(limit)
	page, _ := strconv.Atoi(offset)
	if page > 0 {
		page = (page - 1) * lmt
	}
	return lmt, page
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func FillStruct(m map[string]interface{}, s interface{}) error {
	structValue := reflect.ValueOf(s).Elem()

	for name, value := range m {
		structFieldValue := structValue.FieldByName(name)
		if !structFieldValue.IsValid() {
			return fmt.Errorf("No such field: %s in obj", name)
		}

		if !structFieldValue.CanSet() {
			return fmt.Errorf("Cannot set %s field value", name)
		}

		val := reflect.ValueOf(value)
		if structFieldValue.Type() != val.Type() {
			return errors.New("Provided value type didn't match obj field type")
		}

		structFieldValue.Set(val)
	}
	return nil
}

func RemoveValueRange(hashCode []string, idx int, count int) []string {
	return append(hashCode[:idx], hashCode[idx+count:]...)
}

func ColumnValues(columns []string) (string, string) {

	var params []string
	for i, _ := range columns {
		params = append(params, fmt.Sprintf("$%d", i+1))
	}

	return strings.Join(columns, " , "), strings.Join(params, " , ")
}

func MergeData(t interface{}, u interface{}) {
	mergo.Merge(&t, u)
}

func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

func ComparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		//log.Println(err)
		return false
	}

	return true
}

func GetPtrValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

func GetDatePtrValue(v *time.Time) time.Time {
	if v != nil {
		return *v
	}
	return time.Time{}
}

func GetIntPtrValue(v *int) int {
	if v != nil {
		return *v
	}
	return 0
}

func GetFloatPtrValue(v *float64) float64 {
	if v != nil {
		return *v
	}
	return 0
}

func SetStringPtrValue(v string) *string {
	var va *string
	if v != "" {
		va = &v
	}
	return va
}

func SetIntPtrValue(v int) *int {
	return &v
}

func IsValidStrongPassword(s string) (bool, string) {
	var message string
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 7 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		message = "Harus Mengandung Satu Huruf Kapital "
	}
	if !hasLower {
		message = "Harus Mengandung Satu Huruf Kecil "
	}

	if !hasNumber {
		message = "Harus Mengandung Satu Angka"
	}

	if !hasSpecial {
		message = "Harus Mengandung Satu Karakater Spesial"
	}

	if !hasMinLen {
		message = "Min Password 8 Karakter"
	}

	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial, message
}

func EscapeString(str interface{}) string {
	rexp, _ := regexp.Compile(`'`)
	result := fmt.Sprintf("%v", str)
	result = rexp.ReplaceAllString(result, "''")
	rexp, _ = regexp.Compile(`"`)
	result = rexp.ReplaceAllString(result, `\"`)

	return result
}

func GetEnv(key string) string {
	err := godotenv.Load()
	if err != nil {
		log.Println("Cannot load file .env: ", err)
		panic(err)
	}

	value := GetEnvOrDefault(key, "").(string)
	return value
}

func GetEnvOrDefault(key string, defaultValue interface{}) interface{} {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func ExpectedUint(v interface{}) uint {
	var result uint
	switch v := v.(type) {
	case int:
		result = uint(v)
	case float64:
		result = uint(v)
	case string:
		convertedString, _ := strconv.ParseUint(v, 10, 32)
		result = uint(convertedString)
	}
	return result
}

func ExpectedInt(v interface{}) int {
	var result int
	switch v.(type) {
	case int:
		result = v.(int)
	case float64:
		result = int(v.(float64))
	case string:
		result, _ = strconv.Atoi(v.(string))
	}
	return result
}

func ExpectedString(v interface{}) string {
	var result string
	switch v := v.(type) {
	case int, uint:
		result = fmt.Sprintf("%d", v)
	case float64:
		result = fmt.Sprintf("%f", v)
	case string:
		result = v
	}
	return result
}

func GetNowTime() time.Time {
	return time.Now()
}

func ParseDateStringOnly(datestring string, format string) time.Time {
	layout := "2006-01-02"

	if format != "" {
		layout = format
	}

	t, _ := time.Parse(layout, datestring)
	return t
}

func ParseDateOnly(datestring string, format string) string {
	layout := "2006-01-02"

	if format != "" {
		layout = format
	}

	if datestring == "" {
		return ""
	}
	t, _ := time.Parse(layout, datestring)
	return t.Format("2006-01-02")
}

func DateCheckNil(str string) *string {
	if str != "" && str != "Invalid Date" {
		return &str
	}
	return nil
}

func DateTimeCheckNil(date time.Time) *time.Time {
	if date.IsZero() {
		return nil
	}
	return &date
}

func ParseDateToString(datestring time.Time, format string) string {
	return datestring.Format(format)
}

func Ptr[T any](v T) *T {
	return &v
}

func GetValues(T any) []interface{} {
	v := reflect.ValueOf(T)
	var fields []interface{}
	for i := 0; i < v.NumField(); i++ {
		//excludeTs := slices.Contains(protectedFields, v.Type().Field(i).Tag.Get("db"))
		//if !excludeTs {
		if v.Field(i).Interface() != nil {
			fields = append(fields, v.Field(i).Interface())
		}
	}
	return fields
}

func GetInsertValues(T any) []interface{} {
	v := reflect.ValueOf(T)
	var fields []interface{}
	for i := 0; i < v.NumField(); i++ {
		excludeTs := slices.Contains(protectedFields, v.Type().Field(i).Tag.Get("db"))
		if !IsNilish(v.Field(i).Interface()) {
			if !excludeTs {
				if v.Field(i).Interface() != nil {
					fields = append(fields, v.Field(i).Interface())
				}
			}
		}

	}
	return fields
}

func PrintInsertFields(v interface{}) []string {
	var fields []string
	val := reflect.ValueOf(v)
	for i := 0; i < val.Type().NumField(); i++ {
		exclude := []string{oneToMany, oneToOne, manyToOne}
		valid := slices.Contains(exclude, val.Type().Field(i).Tag.Get("relation"))
		excludeTs := slices.Contains(protectedFields, val.Type().Field(i).Tag.Get("db"))
		if !excludeTs {
			if !valid {
				name := val.Type().Field(i).Tag.Get("db")
				if name != "total_data" {
					fields = append(fields, name)
				}
			}
		}

		//if val.Field(i).Interface() != nil {
		//}
	}

	return fields
}

func PrintFields(v interface{}) []string {
	var fields []string
	val := reflect.ValueOf(v)
	for i := 0; i < val.Type().NumField(); i++ {
		exclude := []string{oneToMany, oneToOne, manyToOne}
		valid := slices.Contains(exclude, val.Type().Field(i).Tag.Get("relation"))
		//excludeTs := slices.Contains(protectedFields, val.Type().Field(i).Tag.Get("db"))
		if !valid {
			name := val.Type().Field(i).Tag.Get("db")
			if name != "total_data" {
				fields = append(fields, name)
			}
		}
		//if val.Field(i).Interface() != nil {
		//}
	}

	return fields
}

func PrintFieldsWithAlias(v interface{}, aliasTable string) []string {
	var fields []string
	val := reflect.ValueOf(v)
	for i := 0; i < val.Type().NumField(); i++ {
		exclude := []string{oneToMany, oneToOne, manyToOne}
		relation := slices.Contains(exclude, val.Type().Field(i).Tag.Get("relation"))
		if !relation {
			name := val.Type().Field(i).Tag.Get("db")
			if name != "total_data" {
				fields = append(fields, fmt.Sprintf(`%s.%s`, aliasTable, name))
			}
		}
	}

	return fields
}

func GetTableName(value interface{}) string {
	if t := reflect.TypeOf(value); t.Kind() == reflect.Ptr {
		return strcase.ToSnake("*" + t.Elem().Name())
	} else {
		return strcase.ToSnake(t.Name())
	}
}

func StructMap(in interface{}) map[string]interface{} {
	v := reflect.ValueOf(in)
	result := make(map[string]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		exclude := []string{oneToMany, oneToOne, manyToOne}
		relation := slices.Contains(exclude, v.Type().Field(i).Tag.Get("relation"))
		if !relation {
			name := v.Type().Field(i).Tag.Get("db")
			if name != "total_data" {
				result[name] = v.Field(i).Interface()
			}
		}
	}

	return result

}

func StrutForScan(u interface{}) []interface{} {
	val := reflect.ValueOf(u).Elem()
	v := make([]interface{}, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		v[i] = valueField.Addr().Interface()
	}
	return v
}

func TypeConverter[R any](data any) (*R, error) {
	var result R
	b, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

func PrintFieldsRelation(v interface{}, aliasTable string) []string {
	var fields []string
	val := reflect.ValueOf(v)
	for i := 0; i < val.Type().NumField(); i++ {
		exclude := []string{oneToMany, oneToOne, manyToOne}
		relation := slices.Contains(exclude, val.Type().Field(i).Tag.Get("relation"))
		if !relation {
			name := val.Type().Field(i).Tag.Get("db")
			fields = append(fields, fmt.Sprintf(`'%s'`, name))
			fields = append(fields, fmt.Sprintf(`%s.%s`, aliasTable, name))
		}
	}
	return fields
}

type DateTimeString string

func (d *DateTimeString) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // Remove quotes around the string
	t, err := time.Parse(DDMMYYYYhhmmss, s)
	if err != nil {
		//return err
	}
	*d = DateTimeString(t.Format(DDMMYYYYhhmmss))
	return nil
}

func (ct *DateTimeString) Scan(value interface{}) error {
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("unexpected type for CustomTime")
	}
	*ct = DateTimeString(t.Format(DDMMYYYYhhmmss))
	return nil
}

func SetExistingFields(src interface{}, dst interface{}) {

	srcFields := reflect.TypeOf(src).Elem()
	srcValues := reflect.ValueOf(src).Elem()

	dstValues := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcFields.NumField(); i++ {
		srcField := srcFields.Field(i)
		srcValue := srcValues.Field(i)

		dstValue := dstValues.FieldByName(srcField.Name)

		if dstValue.IsValid() {
			if dstValue.CanSet() {
				dstValue.Set(srcValue)
			}
		}

	}
}

func DecoderConfig(req interface{}) *mapstructure.DecoderConfig {
	config := &mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &req,
		TagName:     "json",
	}
	return config
}

func IsNilish(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}

	return false
}

func Convert12HourTo24Hour(hour string) string {
	layout := "03:04PM"
	layout24h := "15:04"
	t, err := time.Parse(layout, hour)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return t.Format(layout24h)
}

// take password as input and generate new hash password from it
func GeneratehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// compare plain password with hash password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate JWT token
func GenerateJWT(email, role, key string) (string, error) {
	var mySigningKey = []byte(key)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["email"] = email
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Errorf("Something went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}
