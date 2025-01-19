import grpc
import test_pb2 as check_status_pb2
import test_pb2_grpc as check_status_pb2_grpc

# 设置一个合法的 Authorization header
def run():
    # 连接到gRPC服务器
    channel = grpc.insecure_channel('127.0.0.1:50001')
    stub = check_status_pb2_grpc.checkStatusStub(channel)

    # 设置metadata，传递Authorization token，确保头部键名小写
    metadata = [('authorization', 'Bearer your_token')]  # token替换为实际值

    try:
        # 调用服务并传递metadata
        response = stub.getStatus(check_status_pb2.Empty(), metadata=metadata)
        print("Response received:", response)
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")

if __name__ == "__main__":
    run()
