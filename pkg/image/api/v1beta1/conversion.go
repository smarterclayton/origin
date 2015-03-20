package v1beta1

import (
	"sort"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/conversion"

	newer "github.com/openshift/origin/pkg/image/api"
)

func init() {
	err := kapi.Scheme.AddConversionFuncs(
		// The docker metadata must be cast to a version
		func(in *newer.Image, out *Image, s conversion.Scope) error {
			if err := s.Convert(&in.ObjectMeta, &out.ObjectMeta, 0); err != nil {
				return err
			}

			out.DockerImageReference = in.DockerImageReference
			out.DockerImageManifest = in.DockerImageManifest

			version := in.DockerImageMetadataVersion
			if len(version) == 0 {
				version = "1.0"
			}
			data, err := kapi.Scheme.EncodeToVersion(&in.DockerImageMetadata, version)
			if err != nil {
				return err
			}
			out.DockerImageMetadata.RawJSON = data
			out.DockerImageMetadataVersion = version

			return nil
		},
		func(in *Image, out *newer.Image, s conversion.Scope) error {
			if err := s.Convert(&in.ObjectMeta, &out.ObjectMeta, 0); err != nil {
				return err
			}

			out.DockerImageReference = in.DockerImageReference
			out.DockerImageManifest = in.DockerImageManifest

			version := in.DockerImageMetadataVersion
			if len(version) == 0 {
				version = "1.0"
			}
			if len(in.DockerImageMetadata.RawJSON) > 0 {
				// TODO: add a way to default the expected kind and version of an object if not set
				obj, err := kapi.Scheme.New(version, "DockerImage")
				if err != nil {
					return err
				}
				if err := kapi.Scheme.DecodeInto(in.DockerImageMetadata.RawJSON, obj); err != nil {
					return err
				}
				if err := s.Convert(obj, &out.DockerImageMetadata, 0); err != nil {
					return err
				}
			}
			out.DockerImageMetadataVersion = version

			return nil
		},
		func(in *ImageRepositoryStatus, out *newer.ImageRepositoryStatus, s conversion.Scope) error {
			out.DockerImageRepository = in.DockerImageRepository
			out.Tags = make(map[string]newer.TagEventList)
			return s.Convert(&in.Tags, &out.Tags, 0)
		},
		func(in *newer.ImageRepositoryStatus, out *ImageRepositoryStatus, s conversion.Scope) error {
			out.DockerImageRepository = in.DockerImageRepository
			out.Tags = make([]NamedTagEventList, 0, 0)
			return s.Convert(&in.Tags, &out.Tags, 0)
		},
		func(in *[]NamedTagEventList, out *map[string]newer.TagEventList, s conversion.Scope) error {
			for _, curr := range *in {
				newTagEventList := newer.TagEventList{}
				if err := s.Convert(&curr.Items, &newTagEventList.Items, 0); err != nil {
					return err
				}
				(*out)[curr.Tag] = newTagEventList
			}

			return nil
		},
		func(in *map[string]newer.TagEventList, out *[]NamedTagEventList, s conversion.Scope) error {
			allKeys := make([]string, 0, len(*in))
			for key := range *in {
				allKeys = append(allKeys, key)
			}
			sort.Strings(allKeys)

			for _, key := range allKeys {
				newTagEventList := (*in)[key]
				oldTagEventList := &NamedTagEventList{Tag: key}
				if err := s.Convert(&newTagEventList.Items, &oldTagEventList.Items, 0); err != nil {
					return err
				}

				*out = append(*out, *oldTagEventList)
			}

			return nil
		},
	)
	if err != nil {
		// If one of the conversion functions is malformed, detect it immediately.
		panic(err)
	}
}
