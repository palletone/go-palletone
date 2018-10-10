go install github.com/golang/mock/mockgen
mockgen -source=./dag/interface.go -destination=./dag/dag_mock.go -package=dag