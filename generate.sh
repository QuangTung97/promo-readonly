PROTO_DIR=proto/promo
RPC_DIR=promopb

CURRENT_DIR=$(pwd)
PROTOC_INCLUDES=.:"$CURRENT_DIR/proto/include"

RED='\033[0;31m'
ORANGE='\033[0;33m'
GREEN='\033[0;32m'
NO_COLOR='\033[0m'

generate() {
  echo "=========================================================="
  echo "${ORANGE}Generating: $1${NO_COLOR}"
  echo "----------------------------------------------------------"

  echo $1

  OUTPUT_DIR="$CURRENT_DIR/$RPC_DIR"

  cd "$CURRENT_DIR/$PROTO_DIR" || exit 1

  protoc -I"$PROTOC_INCLUDES" \
    --go_out=paths=source_relative:$OUTPUT_DIR \
    --go-grpc_out=paths=source_relative:$OUTPUT_DIR \
    --grpc-gateway_out=logtostderr=true,paths=source_relative:$OUTPUT_DIR \
    "$1"
  if [ $? -ne 0 ]; then
    echo "----------------------------------------------------------"
    echo "${RED}ERROR while generating: $1${NO_COLOR}"
  else
    echo "${GREEN}Generated: $1${NO_COLOR}"
  fi

  cd "$CURRENT_DIR" || exit 1
}

get_proto_file() {
  basename "$1"
}

PROTO_FILES=$(find $PROTO_DIR -name "*.proto")

rm -rf $RPC_DIR
mkdir -p $RPC_DIR

for file in $PROTO_FILES; do
  generate "$(get_proto_file "$file")"
done
echo "=========================================================="
