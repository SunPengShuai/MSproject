# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

import test_pb2 as test__pb2

GRPC_GENERATED_VERSION = '1.69.0'
GRPC_VERSION = grpc.__version__
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    raise RuntimeError(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in test_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
    )


class checkStatusStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.getStatus = channel.unary_unary(
                '/pb.checkStatus/getStatus',
                request_serializer=test__pb2.Empty.SerializeToString,
                response_deserializer=test__pb2.TestMsg.FromString,
                _registered_method=True)
        self.getStatusA = channel.unary_unary(
                '/pb.checkStatus/getStatusA',
                request_serializer=test__pb2.Empty.SerializeToString,
                response_deserializer=test__pb2.TestMsg.FromString,
                _registered_method=True)
        self.health = channel.unary_unary(
                '/pb.checkStatus/health',
                request_serializer=test__pb2.Empty.SerializeToString,
                response_deserializer=test__pb2.Empty.FromString,
                _registered_method=True)


class checkStatusServicer(object):
    """Missing associated documentation comment in .proto file."""

    def getStatus(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def getStatusA(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def health(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_checkStatusServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'getStatus': grpc.unary_unary_rpc_method_handler(
                    servicer.getStatus,
                    request_deserializer=test__pb2.Empty.FromString,
                    response_serializer=test__pb2.TestMsg.SerializeToString,
            ),
            'getStatusA': grpc.unary_unary_rpc_method_handler(
                    servicer.getStatusA,
                    request_deserializer=test__pb2.Empty.FromString,
                    response_serializer=test__pb2.TestMsg.SerializeToString,
            ),
            'health': grpc.unary_unary_rpc_method_handler(
                    servicer.health,
                    request_deserializer=test__pb2.Empty.FromString,
                    response_serializer=test__pb2.Empty.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'pb.checkStatus', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('pb.checkStatus', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class checkStatus(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def getStatus(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/pb.checkStatus/getStatus',
            test__pb2.Empty.SerializeToString,
            test__pb2.TestMsg.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def getStatusA(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/pb.checkStatus/getStatusA',
            test__pb2.Empty.SerializeToString,
            test__pb2.TestMsg.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def health(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/pb.checkStatus/health',
            test__pb2.Empty.SerializeToString,
            test__pb2.Empty.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)