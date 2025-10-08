// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

#ifndef ORT_WRAPPER_H_
#define ORT_WRAPPER_H_

#include <onnxruntime_c_api.h>

const OrtApi *ort_get_api(const OrtApiBase *base);
OrtStatusPtr ort_create_env(const OrtApi *api, OrtLoggingLevel level,
                            const char *name, OrtEnv **out);
OrtStatusPtr ort_create_session_options(const OrtApi *api,
                                        OrtSessionOptions **out);
OrtStatusPtr ort_create_session(const OrtApi *api, const OrtEnv *env,
                                const char *model_path,
                                const OrtSessionOptions *options,
                                OrtSession **out);
OrtStatusPtr ort_create_memory_info(const OrtApi *api, const char *name,
                                    OrtAllocatorType type, int id,
                                    OrtMemType mem_type, OrtMemoryInfo **out);
OrtStatusPtr ort_create_allocator(const OrtApi *api, const OrtSession *session,
                                  const OrtMemoryInfo *mem_info,
                                  OrtAllocator **out);
OrtStatusPtr ort_session_get_input_count(const OrtApi *api,
                                         const OrtSession *session,
                                         size_t *out);
OrtStatusPtr ort_session_get_input_name(const OrtApi *api,
                                        const OrtSession *session, size_t index,
                                        OrtAllocator *allocator, char **out);
OrtStatusPtr ort_create_tensor_with_data_as_ort_value(
	const OrtApi *api, const OrtMemoryInfo *info, void *p_data,
	size_t p_data_len, const int64_t *shape, size_t shape_len,
	ONNXTensorElementDataType type, OrtValue **out);
OrtStatusPtr ort_get_tensor_mutable_data(const OrtApi *api, OrtValue *value,
                                         void **out);
OrtStatusPtr ort_get_tensor_type_and_shape(const OrtApi *api,
                                           const OrtValue *value,
                                           OrtTensorTypeAndShapeInfo **out);
OrtStatusPtr ort_get_tensor_element_type(const OrtApi *api,
                                         const OrtTensorTypeAndShapeInfo *info,
                                         ONNXTensorElementDataType *out);
OrtStatusPtr ort_get_dimensions_count(const OrtApi *api,
                                      const OrtTensorTypeAndShapeInfo *info,
                                      size_t *out);
OrtStatusPtr ort_get_dimensions(const OrtApi *api,
                                const OrtTensorTypeAndShapeInfo *info,
                                int64_t *dim_values, size_t dim_count);
OrtStatusPtr ort_get_tensor_size_in_bytes(const OrtApi *api,
                                          const OrtValue *value, size_t *size);
OrtStatusPtr ort_run(const OrtApi *api, OrtSession *session,
                     const OrtRunOptions *run_options,
                     const char *const *input_names,
                     const OrtValue *const *inputs, size_t input_len,
                     const char *const *output_names, size_t output_names_len,
                     OrtValue **outputs);
void ort_allocator_free(OrtAllocator *allocator, void *data);
void ort_release_tensor_type_and_shape_info(const OrtApi *api,
                                            OrtTensorTypeAndShapeInfo *info);
void ort_release_status(const OrtApi *api, OrtStatusPtr status);
void ort_release_value(const OrtApi *api, OrtValue *value);
void ort_release_allocator(const OrtApi *api, OrtAllocator *allocator);
void ort_release_memory_info(const OrtApi *api, OrtMemoryInfo *mem_info);
void ort_release_session(const OrtApi *api, OrtSession *session);
void ort_release_session_options(const OrtApi *api, OrtSessionOptions *options);
void ort_release_env(const OrtApi *api, OrtEnv *env);
const char *ort_get_error_message(const OrtApi *api, OrtStatusPtr status);

#endif
