script_dir=$(
    cd $(dirname $0)
    pwd
)                                  # 脚本路径
project_dir=$(dirname $script_dir) # 项目路径

proto_dir=testdata
out_dir=${project_dir}/testdata # 生成代码路径
third_party_dir=../../third_party

protoc \
    -I ${project_dir}/${proto_dir} \
    -I ${third_party_dir} \
    -I ${project_dir} \
    --go_out=${out_dir} \
    --go_opt paths=source_relative \
    registration.proto
