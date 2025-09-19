// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

#include "ort_wrapper.h"

const OrtApi *ort_get_api(const OrtApiBase *base) {
	return base->GetApi(ORT_API_VERSION);
}

OrtStatusPtr ort_create_env(const OrtApi *api, OrtLoggingLevel level,
                            const char *name, OrtEnv **out) {
	return api->CreateEnv(level, name, out);
}

OrtStatusPtr ort_create_session_options(const OrtApi *api,
                                        OrtSessionOptions **out) {
	return api->CreateSessionOptions(out);
}

OrtStatusPtr ort_create_session(const OrtApi *api, const OrtEnv *env,
                                const char *model_path,
                                const OrtSessionOptions *options,
                                OrtSession **out) {
	return api->CreateSession(env, model_path, options, out);
}

OrtStatusPtr ort_create_memory_info(const OrtApi *api, const char *name,
                                    OrtAllocatorType type, int id,
                                    OrtMemType mem_type, OrtMemoryInfo **out) {
	return api->CreateMemoryInfo(name, type, id, mem_type, out);
}

OrtStatusPtr ort_create_allocator(const OrtApi *api, const OrtSession *session,
                                  const OrtMemoryInfo *mem_info,
                                  OrtAllocator **out) {
	return api->CreateAllocator(session, mem_info, out);
}

OrtStatusPtr ort_session_get_input_count(const OrtApi *api,
                                         const OrtSession *session,
                                         size_t *out) {
	return api->SessionGetInputCount(session, out);
}

OrtStatusPtr ort_session_get_input_name(const OrtApi *api,
                                        const OrtSession *session, size_t index,
                                        OrtAllocator *allocator, char **out) {
	return api->SessionGetInputName(session, index, allocator, out);
}

OrtStatusPtr ort_create_tensor_with_data_as_ort_value(
	const OrtApi *api, const OrtMemoryInfo *info, void *p_data,
	size_t p_data_len, const int64_t *shape, size_t shape_len,
	ONNXTensorElementDataType type, OrtValue **out) {
	return api->CreateTensorWithDataAsOrtValue(info, p_data, p_data_len, shape,
	                                           shape_len, type, out);
}

OrtStatusPtr ort_run(const OrtApi *api, OrtSession *session,
                     const OrtRunOptions *run_options,
                     const char *const *input_names,
                     const OrtValue *const *inputs, size_t input_len,
                     const char *const *output_names, size_t output_names_len,
                     OrtValue **outputs) {
	return api->Run(session, run_options, input_names, inputs, input_len,
	                output_names, output_names_len, outputs);
}

void ort_allocator_free(OrtAllocator *allocator, void *data) {
	allocator->Free(allocator, data);
}

void ort_release_status(const OrtApi *api, OrtStatusPtr status) {
	api->ReleaseStatus(status);
}

void ort_release_allocator(const OrtApi *api, OrtAllocator *allocator) {
	api->ReleaseAllocator(allocator);
}

void ort_release_value(const OrtApi *api, OrtValue *value) {
	api->ReleaseValue(value);
}

void ort_release_memory_info(const OrtApi *api, OrtMemoryInfo *mem_info) {
	api->ReleaseMemoryInfo(mem_info);
}

void ort_release_session(const OrtApi *api, OrtSession *session) {
	api->ReleaseSession(session);
}

void ort_release_session_options(const OrtApi *api,
                                 OrtSessionOptions *options) {
	api->ReleaseSessionOptions(options);
}

void ort_release_env(const OrtApi *api, OrtEnv *env) { api->ReleaseEnv(env); }

const char *ort_get_error_message(const OrtApi *api, OrtStatusPtr status) {
	return api->GetErrorMessage(status);
}
