package parsing

import (
	"errors"

	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/ExplorViz/otel-collector/common/attrib"
)

const DatabaseEntityType string = "database"

// A DatabaseEntity represents a database management system (DBMS),
// and potentially a database and / or table within.
type DatabaseEntity struct {
	// Name of the database management system, e.g. "postgresql".
	SystemName string

	// Name of the database within the system.
	DatabaseName string

	// Name of a table within the database.
	TableName string
}

func (d DatabaseEntity) ID() string {
	return "database" + "|" + d.SystemName + "|" + d.DatabaseName + "|" + d.TableName
}

func (d DatabaseEntity) TelemetryKey() string {
	return "database" + "|" + d.SystemName
}

func (d DatabaseEntity) ToAttributes(attrs *pcommon.Map) {
	attrs.PutStr(string(attrib.ExplorVizAttributes.EntityType.Key), DatabaseEntityType)

	// Use system name as service name. Since database calls are usually only instrumented client-side,
	// the resource attributes will not specify the database as its own service.
	attrs.PutStr(string(attrib.ExplorVizAttributes.ServiceName.Key), d.SystemName)

	attrs.PutStr(string(attrib.ExplorVizAttributes.DatabaseSystemName.Key), d.SystemName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.DatabaseName.Key), d.DatabaseName)
	attrs.PutStr(string(attrib.ExplorVizAttributes.DatabaseTableName.Key), d.TableName)
}

// databaseEntityFromAttribs initializes a new [DatabaseEntity] based on the entries of the provided map.
// If the provided map entries are incomplete (meaning a mandatory attribute is missing),
// then a zero-initialized DatabaseEntity and an error is returned.
func databaseEntityFromAttribs(m pcommon.Map) (DatabaseEntity, error) {
	system, ok := m.Get(string(attrib.ExplorVizAttributes.DatabaseSystemName.Key))
	if !ok || system.Str() == "" {
		return DatabaseEntity{}, errors.New("empty or missing string attribute for database system name")
	}

	db, _ := m.Get(string(attrib.ExplorVizAttributes.DatabaseName.Key))

	table, _ := m.Get(string(attrib.ExplorVizAttributes.DatabaseTableName.Key))

	return DatabaseEntity{
		SystemName:   system.Str(),
		DatabaseName: db.Str(),
		TableName:    table.Str(),
	}, nil
}

// ParseDatabaseTelemetry parses telemetry describing database queries by looking for attributes conforming
// to the [OTel semconv database attributes]. For telemetry to be successfully parsed as a database entity,
// it needs to provide:
//   - a name for the database management system (DBMS) used
//
// [OTel semconv database attributes]: https://opentelemetry.io/docs/specs/semconv/registry/attributes/db/
func ParseDatabaseTelemetry(tr attrib.TelemetryReader) (Entity, error) {
	system := tr.StrAttrib(semconv.DBSystemNameKey)
	if system == "" {
		// Look for deprecated attribute as fallback
		system = tr.StrAttrib("db.system")
	}
	if system == "" {
		return DatabaseEntity{}, errors.New("db parser: empty or missing DBMS name attribute")
	}

	db := tr.StrAttrib(semconv.DBNamespaceKey)
	if db == "" {
		// Look for deprecated attribute as fallback
		db = tr.StrAttrib("db.name")
	}

	table := tr.StrAttrib(semconv.DBCollectionNameKey)
	if table == "" {
		// Look for deprecated attribute as fallback
		db = tr.StrAttrib("db.sql.table")
	}

	return DatabaseEntity{
		SystemName:   system,
		DatabaseName: db,
		TableName:    table,
	}, nil
}
