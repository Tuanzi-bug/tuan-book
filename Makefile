.PHONY: mock
mock:
	@mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user.mock.go
	@mockgen -source=internal/service/code.go -package=svcmocks -destination=internal/service/mocks/code.mock.go
	@mockgen -source=internal/repository/code.go -package=repomocks -destination=internal/repository/mocks/code.mock.go
	@mockgen -source=internal/repository/user.go -package=repomocks -destination=internal/repository/mocks/user.mock.go
	@mockgen -source=internal/repository/dao/user.go -package=daomocks -destination=internal/repository/dao/mocks/user.mock.go
	@mockgen -source=internal/repository/cache/user.go -package=cachemocks -destination=internal/repository/cache/mocks/user.mock.go
	@mockgen -source=internal/repository/cache/code.go -package=cachemocks -destination=internal/repository/cache/mocks/code.mock.go
	@mockgen -package=redismocks -destination=internal/repository/cache/redismocks/cmd.mock.go "github.com/redis/go-redis/v9" Cmdable
	@mockgen -source=internal/service/sms/types.go -package=smsmocks -destination=internal/service/sms/mocks/sms.mock.go
	@mockgen -source=pkg/limiter/types.go -package=limitermocks -destination=pkg/limiter/mocks/limiter.mock.go
	@go mod tidy