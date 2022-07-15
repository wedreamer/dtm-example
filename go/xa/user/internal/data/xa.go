package data

import "gorm.io/gorm"

type DbType = uint8
const MYSQL = uint8(1)
const PGSQL = uint8(2)

type XASql struct {
	StartSql    string
	MainSql     string
	EndSql      string
	PrepareSql  string
	CommitSql   string
	RollbackSql string
}

type XA struct {
	Db       *gorm.DB
	Xid      string
	StartXA  func() error
	EndXA    func() error
	Prepare  func() error
	Commit   func() error
	Rollback func() error
	Sql      *XASql
}

func (target *XA) SetXId(xid string) {
	target.Xid = xid
}

func (target *XA) SetStartXA(startXA func() error) {
	target.StartXA = startXA
}

func (target *XA) SetEndXA(endXA func() error) {
	target.EndXA = endXA
}

func (target *XA) SetPrepare(prepare func() error) {
	target.Prepare = prepare
}

func (target *XA) SetCommit(commit func() error) {
	target.Commit = commit
}

func (target *XA) SetRollback(rollback func() error) {
	target.Rollback = rollback
}

func (target *XA) SetSql(mainSql string, dbType DbType) {
	if dbType == MYSQL {
		/* XA start '4fPqCNTYeSG' -- start a xa transaction
		UPDATE `user_account` SET `balance`=balance + 30,`update_time`='2021-06-09 11:50:42.438' WHERE user_id = 1
		XA end '4fPqCNTYeSG'
		-- if connection closed before `prepare`, then the transaction is rolled back automatically
		XA prepare '4fPqCNTYeSG'
		-- When all participants have all prepared, call commit in phase 2
		xa commit '4fPqCNTYeSG'
		-- When any participants have failed to prepare, call rollback in phase 2
		-- xa rollback '4fPqCNTYeSG' */
		target.Sql = &XASql{
			StartSql:    "BEGIN;",
			MainSql:     mainSql,
			EndSql:      "",
			PrepareSql:  T("PREPARE TRANSACTION '%s';", target.Xid),
			CommitSql:   T("COMMIT PREPARED '%s';", target.Xid),
			RollbackSql: T("ROLLBACK PREPARED '%s';", target.Xid),
		}
	} else if dbType == PGSQL {
		/* BEGIN;
		-- DO THINGS TO BE DONE IN A ALL OR NOTHING FASHION
		-- Stop point --
		PREPARE TRANSACTION 't2';

		COMMIT PREPARED 't2' || ROLLBACK PREPARED 't2' */
		target.Sql = &XASql{
			StartSql:    T("XA start '%s';", target.Xid),
			MainSql:     mainSql,
			EndSql:      T("XA end '%s';", target.Xid),
			PrepareSql:  T("XA prepare '%s';", target.Xid),
			CommitSql:   T("XA commit '%s';", target.Xid),
			RollbackSql: T("XA rollback '%s';", target.Xid),
		}
	}
}

func (target *XA) Init(db *gorm.DB, dbType DbType) {
	target.SetStartXA(func() error {
		res := db.Exec(target.Sql.StartSql)
		return res.Error
	})
	if dbType == MYSQL {
		target.SetEndXA(func() error {
			res := db.Exec(target.Sql.MainSql)
			if res.Error != nil {
				return res.Error
			}
			res = db.Exec(target.Sql.EndSql)
			return res.Error
		})
	} else if dbType == PGSQL {
		target.SetEndXA(func() error {
			res := db.Exec(target.Sql.MainSql)
			return res.Error
		})
	}
	target.SetPrepare(func() error {
		res := db.Exec(target.Sql.PrepareSql)
		return res.Error
	})
	target.SetCommit(func() error {
		res := db.Exec(target.Sql.CommitSql)
		return res.Error
	})
	target.SetRollback(func() error {
		res := db.Exec(target.Sql.RollbackSql)
		return res.Error
	})
}
