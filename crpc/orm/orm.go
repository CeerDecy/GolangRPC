package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc/crpcLogger"
	"reflect"
	"strings"
	"time"
)

type CrDB struct {
	db     *sql.DB
	logger *crpcLogger.Logger
}

type CrSession struct {
	db          *CrDB
	TableName   string
	FieldName   []string
	placeHolder []string
	values      []any
}

func (c *CrDB) Close() error {
	return c.db.Close()
}

func Open(driverName, source string) *CrDB {
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	// 最大空闲连接数，默认为2
	db.SetMaxIdleConns(5)
	// 最大连接数，默认不配置是不限制最大连接数
	db.SetMaxOpenConns(100)
	// 最大存活时间
	db.SetConnMaxLifetime(3 * time.Minute)
	// 空闲连接最大存活时间
	db.SetConnMaxIdleTime(time.Minute)
	// 尝试ping一下数据库
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return &CrDB{
		db:     db,
		logger: crpcLogger.TextLogger(),
	}
}

func (c *CrDB) SetMaxIdleConns(n int) {
	c.db.SetMaxIdleConns(n)
}
func (c *CrDB) SetMaxOpenConns(n int) {
	c.db.SetMaxOpenConns(n)
}
func (c *CrDB) SetConnMaxLifetime(n time.Duration) {
	c.db.SetConnMaxLifetime(n)
}
func (c *CrDB) SetConnMaxIdleTime(n time.Duration) {
	c.db.SetConnMaxIdleTime(n)
}

func (c *CrDB) NewSession() *CrSession {
	return &CrSession{
		db: c,
	}
}

func (session *CrSession) Table(name string) *CrSession {
	session.TableName = name
	return session
}

func (session *CrSession) Insert(data any) (int64, int64, error) {
	session.fieldNames(data)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		session.TableName,
		strings.Join(session.FieldName, ","),
		strings.Join(session.placeHolder, ","))
	session.db.logger.Info("["+session.TableName+"]", query)
	stmt, err := session.db.db.Prepare(query)
	if err != nil {
		return -1, -1, err
	}
	r, err := stmt.Exec(session.values...)
	if err != nil {
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (session *CrSession) fieldNames(data any) {
	// 反射
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data must be pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	for i := 0; i < tVar.NumField(); i++ {
		fieldName := tVar.Field(i).Name
		tag := tVar.Field(i).Tag
		sqlTag := tag.Get("corm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(Name(fieldName))
		} else {
			if strings.Contains(sqlTag, "auto_increment") {
				continue
			}
			if strings.Contains(sqlTag, ",") {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
		}
		if sqlTag == "id" && isAutoId(vVar.Field(i).Interface()) {
			continue
		}
		session.FieldName = append(session.FieldName, sqlTag)
		session.placeHolder = append(session.placeHolder, "?")
		session.values = append(session.values, vVar.Field(i).Interface())
	}
}

func isAutoId(id any) bool {
	t := reflect.TypeOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if id.(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if id.(int32) <= 0 {
			return true
		}
	case reflect.Int:
		if id.(int) <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

// Name 添加下划线 UserName -> User_Name
func Name(name string) string {
	var names = name[:]
	lastIndex := 0
	var builder strings.Builder
	for index, value := range names {
		if value >= 'A' && value <= 'Z' {
			if index == 0 {
				continue
			}
			builder.WriteString(names[:index])
			builder.WriteString("_")
			lastIndex = index
		}
	}
	builder.WriteString(names[lastIndex:])
	return builder.String()
}
