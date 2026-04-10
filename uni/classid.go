// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package uni

type ClassID int

const (
	ClassIDUnknown                 ClassID = -1
	ClassIDObject                  ClassID = 0
	ClassIDGameObject              ClassID = 1
	ClassIDTransform               ClassID = 4
	ClassIDCamera                  ClassID = 20
	ClassIDMaterial                ClassID = 21
	ClassIDMeshRenderer            ClassID = 23
	ClassIDTexture2D               ClassID = 28
	ClassIDMeshFilter              ClassID = 33
	ClassIDMesh                    ClassID = 43
	ClassIDShader                  ClassID = 48
	ClassIDTextAsset               ClassID = 49
	ClassIDBoxCollider2D           ClassID = 61
	ClassIDBoxCollider             ClassID = 65
	ClassIDComputeShader           ClassID = 72
	ClassIDAnimationClip           ClassID = 74
	ClassIDAudioSource             ClassID = 82
	ClassIDAudioClip               ClassID = 83
	ClassIDRenderTexture           ClassID = 84
	ClassIDAvatar                  ClassID = 90
	ClassIDAnimatorController      ClassID = 91
	ClassIDAnimator                ClassID = 95
	ClassIDRenderSettings          ClassID = 104
	ClassIDLIGHT                   ClassID = 108
	ClassIDMonoBehaviour           ClassID = 114
	ClassIDMonoScript              ClassID = 115
	ClassIDFont                    ClassID = 128
	ClassIDSphereCollider          ClassID = 135
	ClassIDSkinnedMeshRenderer     ClassID = 137
	ClassIDAssetBundle             ClassID = 142
	ClassIDPreloadData             ClassID = 150
	ClassIDLightmapSettings        ClassID = 157
	ClassIDNavMeshSettings         ClassID = 196
	ClassIDParticleSystem          ClassID = 198
	ClassIDParticleSystemRenderer  ClassID = 199
	ClassIDShaderVariantCollection ClassID = 200
	ClassIDSpriteRenderer          ClassID = 212
	ClassIDSprite                  ClassID = 213
	ClassIDCanvasRenderer          ClassID = 222
	ClassIDCanvas                  ClassID = 223
	ClassIDRectTransform           ClassID = 224
	ClassIDCanvasGroup             ClassID = 225
	ClassIDPlayableDirector        ClassID = 320
)
