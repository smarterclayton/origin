package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/docker/distribution/digest"
)

// DockerDefaultNamespace is the value for namespace when a single segment name is provided.
const DockerDefaultNamespace = "library"

// DockerImageReference points to a Docker image.
type DockerImageReference struct {
	Registry  string
	Namespace string
	Name      string
	Tag       string
	ID        string
}

// TODO remove (base, tag, id)
func parseRepositoryTag(repos string) (string, string, string) {
	n := strings.Index(repos, "@")
	if n >= 0 {
		parts := strings.Split(repos, "@")
		return parts[0], "", parts[1]
	}
	n = strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, "", ""
	}
	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag, ""
	}
	return repos, "", ""
}

// ParseDockerImageReference parses a Docker pull spec string into a
// DockerImageReference.
func ParseDockerImageReference(spec string) (DockerImageReference, error) {
	var (
		ref DockerImageReference
	)
	// TODO replace with docker version once docker/docker PR11109 is merged upstream
	repo, tag, id := parseRepositoryTag(spec)

	repoParts := strings.Split(repo, "/")
	switch len(repoParts) {
	case 2:
		// namespace/name
		ref.Namespace = repoParts[0]
		ref.Name = repoParts[1]
		ref.Tag = tag
		ref.ID = id
		return ref, nil
	case 3:
		// registry/namespace/name
		ref.Registry = repoParts[0]
		ref.Namespace = repoParts[1]
		ref.Name = repoParts[2]
		ref.Tag = tag
		ref.ID = id
		return ref, nil
	case 1:
		// name
		if len(repoParts[0]) == 0 {
			return ref, fmt.Errorf("the docker pull spec %q must be two or three segments separated by slashes", spec)
		}
		ref.Name = repoParts[0]
		ref.Tag = tag
		ref.ID = id
		return ref, nil
	default:
		return ref, fmt.Errorf("the docker pull spec %q must be two or three segments separated by slashes", spec)
	}
}

// DockerClientDefaults sets the default values used by the Docker client.
func (r DockerImageReference) DockerClientDefaults() DockerImageReference {
	if len(r.Namespace) == 0 {
		r.Namespace = "library"
	}
	if len(r.Registry) == 0 {
		r.Registry = "index.docker.io"
	}
	if len(r.Tag) == 0 {
		r.Tag = "latest"
	}
	return r
}

var dockerPullSpecGenerator pullSpecGenerator

// String converts a DockerImageReference to a Docker pull spec.
func (r DockerImageReference) String() string {
	if dockerPullSpecGenerator == nil {
		if len(os.Getenv("OPENSHIFT_REAL_PULL_BY_ID")) > 0 {
			dockerPullSpecGenerator = &realByIdPullSpecGenerator{}
		} else {
			dockerPullSpecGenerator = &simulatedByIdPullSpecGenerator{}
		}
	}
	return dockerPullSpecGenerator.pullSpec(r)
}

// pullSpecGenerator converts a DockerImageReference to a Docker pull spec.
type pullSpecGenerator interface {
	pullSpec(ref DockerImageReference) string
}

// simulatedByIdPullSpecGenerator simulates pull by ID against a v2 registry
// by generating a pull spec where the "tag" is the hex portion of the
// DockerImageReference's ID.
type simulatedByIdPullSpecGenerator struct{}

func (f *simulatedByIdPullSpecGenerator) pullSpec(r DockerImageReference) string {
	registry := r.Registry
	if len(registry) > 0 {
		registry += "/"
	}

	if len(r.Namespace) == 0 {
		r.Namespace = DockerDefaultNamespace
	}
	r.Namespace += "/"

	var ref string
	if len(r.Tag) > 0 {
		ref = ":" + r.Tag
	} else if len(r.ID) > 0 {
		if d, err := digest.ParseDigest(r.ID); err == nil {
			// if it parses as a digest, treat it like a by-id tag without the algorithm
			ref = ":" + d.Hex()
		} else {
			// if it doesn't parse, it's presumably a v1 registry by-id tag
			ref = ":" + r.ID
		}
	}

	return fmt.Sprintf("%s%s%s%s", registry, r.Namespace, r.Name, ref)
}

// realByIdPullSpecGenerator generates real pull by ID pull specs against
// a v2 registry using the <repo>@<algo:digest> format.
type realByIdPullSpecGenerator struct{}

func (*realByIdPullSpecGenerator) pullSpec(r DockerImageReference) string {
	registry := r.Registry
	if len(registry) > 0 {
		registry += "/"
	}

	if len(r.Namespace) == 0 {
		r.Namespace = DockerDefaultNamespace
	}
	r.Namespace += "/"

	var ref string
	if len(r.Tag) > 0 {
		ref = ":" + r.Tag
	} else if len(r.ID) > 0 {
		ref = "@" + r.ID
	}

	return fmt.Sprintf("%s%s%s%s", registry, r.Namespace, r.Name, ref)
}

// ImageWithMetadata returns a copy of image with the DockerImageMetadata filled in
// from the raw DockerImageManifest data stored in the image.
func ImageWithMetadata(image Image) (*Image, error) {
	if len(image.DockerImageManifest) == 0 {
		return &image, nil
	}

	manifestData := image.DockerImageManifest

	image.DockerImageManifest = ""

	manifest := DockerImageManifest{}
	if err := json.Unmarshal([]byte(manifestData), &manifest); err != nil {
		return nil, err
	}

	if len(manifest.History) == 0 {
		// should never have an empty history, but just in case...
		return &image, nil
	}

	v1Metadata := DockerV1CompatibilityImage{}
	if err := json.Unmarshal([]byte(manifest.History[0].DockerV1Compatibility), &v1Metadata); err != nil {
		return nil, err
	}

	image.DockerImageMetadata.ID = v1Metadata.ID
	image.DockerImageMetadata.Parent = v1Metadata.Parent
	image.DockerImageMetadata.Comment = v1Metadata.Comment
	image.DockerImageMetadata.Created = v1Metadata.Created
	image.DockerImageMetadata.Container = v1Metadata.Container
	image.DockerImageMetadata.ContainerConfig = v1Metadata.ContainerConfig
	image.DockerImageMetadata.DockerVersion = v1Metadata.DockerVersion
	image.DockerImageMetadata.Author = v1Metadata.Author
	image.DockerImageMetadata.Config = v1Metadata.Config
	image.DockerImageMetadata.Architecture = v1Metadata.Architecture
	image.DockerImageMetadata.Size = v1Metadata.Size

	return &image, nil
}

func TagValueToTagEvent(repo *ImageRepository, value string) (*TagEvent, error) {
	if strings.Contains(value, "@") {
		segs := strings.SplitN(value, "@", 2)
		if len(segs[1]) == 0 {
			return nil, fmt.Errorf("%q may not end with a @", value)
		}
		ref, err := DockerImageReferenceForRepository(repo)
		if err != nil {
			return nil, err
		}
		ref.ID = segs[1]
		return &TagEvent{
			Created:              util.Now(),
			DockerImageReference: ref.String(),
			Image:                ref.ID,
		}, nil
	}
	return LatestTaggedImage(repo, value)
}

// DockerImageReferenceForRepository returns a DockerImageReference that represents
// the ImageRepository or false, if no valid reference exists.
func DockerImageReferenceForRepository(repo *ImageRepository) (DockerImageReference, error) {
	spec := repo.Status.DockerImageRepository
	if len(spec) == 0 {
		spec = repo.DockerImageRepository
	}
	if len(spec) == 0 {
		return DockerImageReference{}, fmt.Errorf("no possible pull spec for %s/%s", repo.Namespace, repo.Name)
	}
	return ParseDockerImageReference(spec)
}

// LatestTaggedImage returns the most recent TagEvent for the specified image
// repository and tag. Will resolve lookups for the empty tag.
func LatestTaggedImage(repo *ImageRepository, tag string) (*TagEvent, error) {
	if len(tag) == 0 {
		tag = "latest"
	}
	// find the most recent tag event with an image reference
	if repo.Status.Tags != nil {
		if history, ok := repo.Status.Tags[tag]; ok {
			for _, item := range history.Items {
				if len(item.DockerImageReference) > 0 {
					return &item, nil
				}
			}
		}
	}

	// infer a pull spec given the pull locations - requires the tag
	// to have a value, for .status.DIR or .DIR to be set, and for
	// one of those values to be valid.
	if value, ok := repo.Tags[tag]; ok && len(value) > 0 {
		ref, err := DockerImageReferenceForRepository(repo)
		if err != nil {
			return nil, err
		}
		ref.Tag = value
		return &TagEvent{
			Created:              util.Now(),
			DockerImageReference: ref.String(),
		}, nil
	}

	return nil, fmt.Errorf("no image recorded for %s/%s:%s", repo.Namespace, repo.Name, tag)
}

// AddTagEventToImageRepository attempts to update the given image repository with a tag event. It will
// collapse duplicate entries - returning true if a change was made or false if no change
// occurred.
func AddTagEventToImageRepository(repo *ImageRepository, tag string, next TagEvent) bool {
	if repo.Status.Tags == nil {
		repo.Status.Tags = make(map[string]TagEventList)
	}

	tags, ok := repo.Status.Tags[tag]
	if !ok || len(tags.Items) == 0 {
		repo.Status.Tags[tag] = TagEventList{Items: []TagEvent{next}}
		return true
	}

	previous := &tags.Items[0]

	// image reference has not changed
	if previous.DockerImageReference == next.DockerImageReference {
		if next.Image == previous.Image {
			return false
		}
		previous.Image = next.Image
		repo.Status.Tags[tag] = tags
		return true
	}

	// image has not changed, but image reference has
	if next.Image == previous.Image {
		previous.DockerImageReference = next.DockerImageReference
		repo.Status.Tags[tag] = tags
		return true
	}

	tags.Items = append([]TagEvent{next}, tags.Items...)
	repo.Status.Tags[tag] = tags
	return true
}
