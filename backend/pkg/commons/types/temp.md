protoc -I /usr/local/include \
-I . \
--gotag_out=auto="ch-as-snake_case":. types/eth1.proto
