# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: test.proto
# Protobuf Python Version: 5.29.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    0,
    '',
    'test.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\ntest.proto\x12\x02pb\"\x07\n\x05\x45mpty\"&\n\x07TestMsg\x12\x0b\n\x03msg\x18\x01 \x01(\t\x12\x0e\n\x06status\x18\x02 \x01(\x05\x32~\n\x0b\x63heckStatus\x12%\n\tgetStatus\x12\t.pb.Empty\x1a\x0b.pb.TestMsg\"\x00\x12&\n\ngetStatusA\x12\t.pb.Empty\x1a\x0b.pb.TestMsg\"\x00\x12 \n\x06health\x12\t.pb.Empty\x1a\t.pb.Empty\"\x00\x42\x06Z\x04./pbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'test_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\004./pb'
  _globals['_EMPTY']._serialized_start=18
  _globals['_EMPTY']._serialized_end=25
  _globals['_TESTMSG']._serialized_start=27
  _globals['_TESTMSG']._serialized_end=65
  _globals['_CHECKSTATUS']._serialized_start=67
  _globals['_CHECKSTATUS']._serialized_end=193
# @@protoc_insertion_point(module_scope)