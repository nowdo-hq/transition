package transition

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/qor/admin"
	"github.com/qor/audited"
	"github.com/qor/qor/resource"
	"github.com/qor/roles"
	"gorm.io/gorm"
)

// StateChangeLog a model that used to keep state change logs
type StateChangeLog struct {
	ID         uint `gorm:"primarykey"`
	ReferTable string
	ReferID    string
	From       string
	To         string
	Note       string `sql:"size:1024"`
	CreatedAt  int64
	UpdatedAt  int64
	DeletedAt  *int64 `sql:"index"`
	audited.AuditedModel
}

// GenerateReferenceKey generate reference key used for change log
func GenerateReferenceKey(model interface{}, scope *gorm.DB) string {
	var primaryValues = []string{}
	for _, field := range scope.Statement.Schema.Fields {
		if !field.PrimaryKey {
			continue
		}
		var value, _ = field.ValueOf(reflect.ValueOf(model))
		primaryValues = append(primaryValues, fmt.Sprint(value))
	}

	return strings.Join(primaryValues, "::")
}

// GetStateChangeLogs get state change logs
func GetStateChangeLogs(model interface{}, db *gorm.DB) []StateChangeLog {
	var (
		changelogs []StateChangeLog
		scope      = db.Model(model)
	)

	scope.Statement.Parse(model)
	db.Where("refer_table = ? AND refer_id = ?", scope.Statement.Table, GenerateReferenceKey(model, scope)).Find(&changelogs)

	return changelogs
}

// GetLastStateChange gets last state change
func GetLastStateChange(model interface{}, db *gorm.DB) *StateChangeLog {
	var (
		changelog StateChangeLog
		scope     = db.Model(model)
	)

	scope.Statement.Parse(model)
	db.Where("refer_table = ? AND refer_id = ?", scope.Statement.Table, GenerateReferenceKey(model, scope)).Last(&changelog)
	if changelog.To == "" {
		return nil
	}
	return &changelog
}

// ConfigureQorResource used to configure transition for qor admin
func (stageChangeLog *StateChangeLog) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		if res.Permission == nil {
			res.Permission = roles.Deny(roles.Update, roles.Anyone).Deny(roles.Create, roles.Anyone)
		} else {
			res.Permission = res.Permission.Deny(roles.Update, roles.Anyone).Deny(roles.Create, roles.Anyone)
		}
	}
}
