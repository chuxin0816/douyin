// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"douyin/src/dal/model"
)

func newUserLogin(db *gorm.DB, opts ...gen.DOOption) userLogin {
	_userLogin := userLogin{}

	_userLogin.userLoginDo.UseDB(db, opts...)
	_userLogin.userLoginDo.UseModel(&model.UserLogin{})

	tableName := _userLogin.userLoginDo.TableName()
	_userLogin.ALL = field.NewAsterisk(tableName)
	_userLogin.ID = field.NewInt64(tableName, "id")
	_userLogin.Username = field.NewString(tableName, "username")
	_userLogin.Password = field.NewString(tableName, "password")
	_userLogin.CreateTime = field.NewTime(tableName, "create_time")
	_userLogin.UpdateTime = field.NewTime(tableName, "update_time")

	_userLogin.fillFieldMap()

	return _userLogin
}

type userLogin struct {
	userLoginDo userLoginDo

	ALL        field.Asterisk
	ID         field.Int64
	Username   field.String // 用户名
	Password   field.String // 加密密码
	CreateTime field.Time   // 创建时间
	UpdateTime field.Time   // 更新时间

	fieldMap map[string]field.Expr
}

func (u userLogin) Table(newTableName string) *userLogin {
	u.userLoginDo.UseTable(newTableName)
	return u.updateTableName(newTableName)
}

func (u userLogin) As(alias string) *userLogin {
	u.userLoginDo.DO = *(u.userLoginDo.As(alias).(*gen.DO))
	return u.updateTableName(alias)
}

func (u *userLogin) updateTableName(table string) *userLogin {
	u.ALL = field.NewAsterisk(table)
	u.ID = field.NewInt64(table, "id")
	u.Username = field.NewString(table, "username")
	u.Password = field.NewString(table, "password")
	u.CreateTime = field.NewTime(table, "create_time")
	u.UpdateTime = field.NewTime(table, "update_time")

	u.fillFieldMap()

	return u
}

func (u *userLogin) WithContext(ctx context.Context) *userLoginDo {
	return u.userLoginDo.WithContext(ctx)
}

func (u userLogin) TableName() string { return u.userLoginDo.TableName() }

func (u userLogin) Alias() string { return u.userLoginDo.Alias() }

func (u userLogin) Columns(cols ...field.Expr) gen.Columns { return u.userLoginDo.Columns(cols...) }

func (u *userLogin) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := u.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (u *userLogin) fillFieldMap() {
	u.fieldMap = make(map[string]field.Expr, 5)
	u.fieldMap["id"] = u.ID
	u.fieldMap["username"] = u.Username
	u.fieldMap["password"] = u.Password
	u.fieldMap["create_time"] = u.CreateTime
	u.fieldMap["update_time"] = u.UpdateTime
}

func (u userLogin) clone(db *gorm.DB) userLogin {
	u.userLoginDo.ReplaceConnPool(db.Statement.ConnPool)
	return u
}

func (u userLogin) replaceDB(db *gorm.DB) userLogin {
	u.userLoginDo.ReplaceDB(db)
	return u
}

type userLoginDo struct{ gen.DO }

func (u userLoginDo) Debug() *userLoginDo {
	return u.withDO(u.DO.Debug())
}

func (u userLoginDo) WithContext(ctx context.Context) *userLoginDo {
	return u.withDO(u.DO.WithContext(ctx))
}

func (u userLoginDo) ReadDB() *userLoginDo {
	return u.Clauses(dbresolver.Read)
}

func (u userLoginDo) WriteDB() *userLoginDo {
	return u.Clauses(dbresolver.Write)
}

func (u userLoginDo) Session(config *gorm.Session) *userLoginDo {
	return u.withDO(u.DO.Session(config))
}

func (u userLoginDo) Clauses(conds ...clause.Expression) *userLoginDo {
	return u.withDO(u.DO.Clauses(conds...))
}

func (u userLoginDo) Returning(value interface{}, columns ...string) *userLoginDo {
	return u.withDO(u.DO.Returning(value, columns...))
}

func (u userLoginDo) Not(conds ...gen.Condition) *userLoginDo {
	return u.withDO(u.DO.Not(conds...))
}

func (u userLoginDo) Or(conds ...gen.Condition) *userLoginDo {
	return u.withDO(u.DO.Or(conds...))
}

func (u userLoginDo) Select(conds ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.Select(conds...))
}

func (u userLoginDo) Where(conds ...gen.Condition) *userLoginDo {
	return u.withDO(u.DO.Where(conds...))
}

func (u userLoginDo) Order(conds ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.Order(conds...))
}

func (u userLoginDo) Distinct(cols ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.Distinct(cols...))
}

func (u userLoginDo) Omit(cols ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.Omit(cols...))
}

func (u userLoginDo) Join(table schema.Tabler, on ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.Join(table, on...))
}

func (u userLoginDo) LeftJoin(table schema.Tabler, on ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.LeftJoin(table, on...))
}

func (u userLoginDo) RightJoin(table schema.Tabler, on ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.RightJoin(table, on...))
}

func (u userLoginDo) Group(cols ...field.Expr) *userLoginDo {
	return u.withDO(u.DO.Group(cols...))
}

func (u userLoginDo) Having(conds ...gen.Condition) *userLoginDo {
	return u.withDO(u.DO.Having(conds...))
}

func (u userLoginDo) Limit(limit int) *userLoginDo {
	return u.withDO(u.DO.Limit(limit))
}

func (u userLoginDo) Offset(offset int) *userLoginDo {
	return u.withDO(u.DO.Offset(offset))
}

func (u userLoginDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *userLoginDo {
	return u.withDO(u.DO.Scopes(funcs...))
}

func (u userLoginDo) Unscoped() *userLoginDo {
	return u.withDO(u.DO.Unscoped())
}

func (u userLoginDo) Create(values ...*model.UserLogin) error {
	if len(values) == 0 {
		return nil
	}
	return u.DO.Create(values)
}

func (u userLoginDo) CreateInBatches(values []*model.UserLogin, batchSize int) error {
	return u.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (u userLoginDo) Save(values ...*model.UserLogin) error {
	if len(values) == 0 {
		return nil
	}
	return u.DO.Save(values)
}

func (u userLoginDo) First() (*model.UserLogin, error) {
	if result, err := u.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserLogin), nil
	}
}

func (u userLoginDo) Take() (*model.UserLogin, error) {
	if result, err := u.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserLogin), nil
	}
}

func (u userLoginDo) Last() (*model.UserLogin, error) {
	if result, err := u.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserLogin), nil
	}
}

func (u userLoginDo) Find() ([]*model.UserLogin, error) {
	result, err := u.DO.Find()
	return result.([]*model.UserLogin), err
}

func (u userLoginDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.UserLogin, err error) {
	buf := make([]*model.UserLogin, 0, batchSize)
	err = u.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (u userLoginDo) FindInBatches(result *[]*model.UserLogin, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return u.DO.FindInBatches(result, batchSize, fc)
}

func (u userLoginDo) Attrs(attrs ...field.AssignExpr) *userLoginDo {
	return u.withDO(u.DO.Attrs(attrs...))
}

func (u userLoginDo) Assign(attrs ...field.AssignExpr) *userLoginDo {
	return u.withDO(u.DO.Assign(attrs...))
}

func (u userLoginDo) Joins(fields ...field.RelationField) *userLoginDo {
	for _, _f := range fields {
		u = *u.withDO(u.DO.Joins(_f))
	}
	return &u
}

func (u userLoginDo) Preload(fields ...field.RelationField) *userLoginDo {
	for _, _f := range fields {
		u = *u.withDO(u.DO.Preload(_f))
	}
	return &u
}

func (u userLoginDo) FirstOrInit() (*model.UserLogin, error) {
	if result, err := u.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserLogin), nil
	}
}

func (u userLoginDo) FirstOrCreate() (*model.UserLogin, error) {
	if result, err := u.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.UserLogin), nil
	}
}

func (u userLoginDo) FindByPage(offset int, limit int) (result []*model.UserLogin, count int64, err error) {
	result, err = u.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = u.Offset(-1).Limit(-1).Count()
	return
}

func (u userLoginDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = u.Count()
	if err != nil {
		return
	}

	err = u.Offset(offset).Limit(limit).Scan(result)
	return
}

func (u userLoginDo) Scan(result interface{}) (err error) {
	return u.DO.Scan(result)
}

func (u userLoginDo) Delete(models ...*model.UserLogin) (result gen.ResultInfo, err error) {
	return u.DO.Delete(models)
}

func (u *userLoginDo) withDO(do gen.Dao) *userLoginDo {
	u.DO = *do.(*gen.DO)
	return u
}
