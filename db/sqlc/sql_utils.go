package db

import "database/sql"

func Int64ToSqlInt64(inputInt int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: inputInt,
		Valid: true,
	}
}
