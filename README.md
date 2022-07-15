# dtm-example

简体中文 | English

使用 dtm 完成分布式事务的示例, 参考 [dtm 文档](https://www.dtm.pub/)

## go

```go
XaGrpcClient = dtmgrpc.NewXaGrpcClient(DtmGrpcServer, config.DB, BusiGrpc+"/busi.Busi/XaNotify")

gid := dtmgrpc.MustGenGid(DtmGrpcServer)
busiData := dtmcli.MustMarshal(&TransReq{Amount: 30})
err := XaGrpcClient.XaGlobalTransaction(gid, func(xa *dtmgrpc.XaGrpc) error {
	_, err := xa.CallBranch(busiData, BusiGrpc+"/busi.Busi/TransOutXa")
	if err != nil {
		return err
	}
	_, err = xa.CallBranch(busiData, BusiGrpc+"/busi.Busi/TransInXa")
	return err
})

func (s *busiServer) XaNotify(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	err := XaGrpcClient.HandleCallback(in.Info.Gid, in.Info.BranchID, in.Info.BranchType)
	return &emptypb.Empty{}, dtmgrpc.Result2Error(nil, err)
}

func (s *busiServer) TransInXa(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		if req.TransInResult == "FAILURE" {
			return status.New(codes.Aborted, "user return failure").Err()
		}
		_, err := dtmcli.SdbExec(db, "update dtm_busi.user_account set balance=balance+? where user_id=?", req.Amount, 2)
		return err
	})
}

func (s *busiServer) TransOutXa(ctx context.Context, in *dtmgrpc.BusiRequest) (*emptypb.Empty, error) {
	req := TransReq{}
	dtmcli.MustUnmarshal(in.BusiData, &req)
	return &emptypb.Empty{}, XaGrpcClient.XaLocalTransaction(in, func(db *sql.DB, xa *dtmgrpc.XaGrpc) error {
		if req.TransOutResult == "FAILURE" {
			return status.New(codes.Aborted, "user return failure").Err()
		}
		_, err := dtmcli.SdbExec(db, "update dtm_busi.user_account set balance=balance-? where user_id=?", req.Amount, 1)
		return err
	})
}
```

## .net

## node


