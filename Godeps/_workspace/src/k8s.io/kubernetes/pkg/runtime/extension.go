/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runtime

import (
	"encoding/json"
	"errors"

	"k8s.io/kubernetes/pkg/api/unversioned"
)

// NestedObjectEncoder is an optional interface that objects may implement to be given
// an opportunity to encode any nested Objects / RawExtensions during serialization.
// TODO: this would be better implemented if all serializer libraries in go (json, protobuf)
//   were capable of passing a context object during marshal, that would allow objects that
//   contained nested objects to get access to the encoder that was handling the top level.
type NestedObjectEncoder interface {
	EncodeNestedObjects(e Encoder, overrides ...unversioned.GroupVersion) error
}

// NestedObjectDecoder is an optional interface that objects may implement to be given
// an opportunity to decode any nested Objects / RawExtensions during serialization.
// TODO: this would be better implemented if all serializer libraries in go (json, protobuf)
//   were capable of passing a context object during marshal, that would allow objects that
//   contained nested objects to get access to the encoder that was handling the top level.
type NestedObjectDecoder interface {
	DecodeNestedObjects(d Decoder, overrides ...unversioned.GroupVersion) error
}

func (re *RawExtension) UnmarshalJSON(in []byte) error {
	if re == nil {
		return errors.New("runtime.RawExtension: UnmarshalJSON on nil pointer")
	}
	re.RawJSON = append(re.RawJSON[0:0], in...)
	return nil
}

// Marshal may get called on pointers or values, so implement MarshalJSON on value.
// http://stackoverflow.com/questions/21390979/custom-marshaljson-never-gets-called-in-go
func (re RawExtension) MarshalJSON() ([]byte, error) {
	if re.RawJSON == nil {
		// TODO: this is to support legacy behavior of JSONPrinter and YAMLPrinter, which
		// expect to call json.Marshal on arbitrary versioned objects (even those not in
		// the scheme). pkg/kubectl/resource#AsVersionedObjects and its interaction with
		// kubectl get on objects not in the scheme needs to be updated to ensure that the
		// objects that are not part of the scheme are correctly put into the right form.
		if re.Object != nil {
			return json.Marshal(re.Object)
		}
		return []byte("null"), nil
	}
	return re.RawJSON, nil
}
