package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	admissionregv1alpha1 "k8s.io/api/admissionregistration/v1alpha1"
	admissionregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apidiscoveryv2 "k8s.io/api/apidiscovery/v2"
	apidiscoveryv2beta1 "k8s.io/api/apidiscovery/v2beta1"
	apiserverinternalv1alpha1 "k8s.io/api/apiserverinternal/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	authenticationv1 "k8s.io/api/authentication/v1"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	authorizationv1 "k8s.io/api/authorization/v1"
	authorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1alpha1 "k8s.io/api/certificates/v1alpha1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	coordinationv1 "k8s.io/api/coordination/v1"
	coordinationv1beta1 "k8s.io/api/coordination/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	eventsv1 "k8s.io/api/events/v1"
	eventsv1beta1 "k8s.io/api/events/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	flowcontrolv1beta1 "k8s.io/api/flowcontrol/v1beta1"
	flowcontrolv1beta2 "k8s.io/api/flowcontrol/v1beta2"
	flowcontrolv1beta3 "k8s.io/api/flowcontrol/v1beta3"
	imagepolicyv1alpha1 "k8s.io/api/imagepolicy/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1alpha1 "k8s.io/api/networking/v1alpha1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	nodev1 "k8s.io/api/node/v1"
	nodev1alpha1 "k8s.io/api/node/v1alpha1"
	nodev1beta1 "k8s.io/api/node/v1beta1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	schedulingv1 "k8s.io/api/scheduling/v1"
	schedulingv1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	svmv1alpha1 "k8s.io/api/storagemigration/v1alpha1"

	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func isIgnoredType(typ reflect.Type) bool {
	ignoredTypes := []reflect.Type{
		reflect.TypeOf(meta.Time{}),
		reflect.TypeOf(meta.MicroTime{}),
		reflect.TypeOf(time.Time{}),
		reflect.TypeOf(intstr.IntOrString{}),
		reflect.TypeOf(resource.Quantity{}),
	}

	for _, ignoredType := range ignoredTypes {
		if typ == ignoredType {
			return true
		}
	}

	return false
}

type Schema struct {
	Type                 string            `json:"type"`
	Required             []string          `json:"required,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	AdditionalProperties bool              `json:"additionalProperties,omitempty"`
	Arguments            []Schema          `json:"arguments,omitempty"`
	Returns              *Schema           `json:"returns,omitempty"`
	Description          string            `json:"description,omitempty"`
	Default              any               `json:"default,omitempty"`
	PatchStrategy        string            `json:"patchStrategy,omitempty"`
	PatchMergeKey        string            `json:"patchMergeKey,omitempty"`
}

func toSchema(typ reflect.Type) Schema {

	if typ.Kind() == reflect.Ptr {
		return toSchema(typ.Elem())
	}

	if isIgnoredType(typ) {
		return Schema{
			Type: typ.Name(),
		}
	}

	typeMap := map[reflect.Kind]string{
		reflect.Bool:    "boolean",
		reflect.Int:     "integer",
		reflect.Int8:    "integer",
		reflect.Int16:   "integer",
		reflect.Int32:   "integer",
		reflect.Int64:   "integer",
		reflect.Uint:    "integer",
		reflect.Uint8:   "integer",
		reflect.Uint16:  "integer",
		reflect.Uint32:  "integer",
		reflect.Uint64:  "integer",
		reflect.Float32: "number",
		reflect.Float64: "number",
		reflect.String:  "string",
		reflect.Struct:  "object",
		reflect.Map:     "object",
		reflect.Slice:   "array",
		reflect.Array:   "array",
	}

	//var schema Schema
	schema := Schema{
		Type:       typeMap[typ.Kind()],
		Required:   make([]string, 0),
		Properties: make(map[string]Schema),
	}

	if schema.Type == "" {
		fmt.Printf("!!! Unsupported type: %s\n", typ.Kind())
		return schema
	}

	if typ.Kind() == reflect.Struct {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				fmt.Printf("Skipping field %s as it has no json tag\n", field.Name)
				fmt.Printf("%v\n", field)
				continue
			}
			split := strings.Split(jsonTag, ",")
			fieldName := split[0]
			if fieldName == "" || fieldName == "-" {
				continue
			}

			if !strings.Contains(field.Tag.Get("json"), "omitempty") {
				schema.Required = append(schema.Required, fieldName)
			}

			schema.Properties[fieldName] = toSchema(field.Type)
			schema.PatchStrategy = field.Tag.Get("patchStrategy")
			schema.PatchMergeKey = field.Tag.Get("patchMergeKey")
		}
	} else if typ.Kind() == reflect.Map {
		schema.Type = "object"
		schema.AdditionalProperties = true
		elemSchema := toSchema(typ.Elem())
		schema.Items = &elemSchema
	} else if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		itemSchema := toSchema(typ.Elem())
		schema.Items = &itemSchema
	}

	return schema
}

func JsonPrint(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func main() {
	var groups = []runtime.SchemeBuilder{
		admissionv1beta1.SchemeBuilder,
		admissionv1.SchemeBuilder,
		admissionregv1alpha1.SchemeBuilder,
		admissionregv1beta1.SchemeBuilder,
		admissionregv1.SchemeBuilder,
		apiserverinternalv1alpha1.SchemeBuilder,
		apidiscoveryv2beta1.SchemeBuilder,
		apidiscoveryv2.SchemeBuilder,
		appsv1beta1.SchemeBuilder,
		appsv1beta2.SchemeBuilder,
		appsv1.SchemeBuilder,
		authenticationv1beta1.SchemeBuilder,
		authenticationv1.SchemeBuilder,
		authorizationv1beta1.SchemeBuilder,
		authorizationv1.SchemeBuilder,
		autoscalingv1.SchemeBuilder,
		autoscalingv2.SchemeBuilder,
		autoscalingv2beta1.SchemeBuilder,
		autoscalingv2beta2.SchemeBuilder,
		batchv1beta1.SchemeBuilder,
		batchv1.SchemeBuilder,
		certificatesv1.SchemeBuilder,
		certificatesv1beta1.SchemeBuilder,
		certificatesv1alpha1.SchemeBuilder,
		coordinationv1.SchemeBuilder,
		coordinationv1beta1.SchemeBuilder,
		corev1.SchemeBuilder,
		discoveryv1.SchemeBuilder,
		discoveryv1beta1.SchemeBuilder,
		eventsv1.SchemeBuilder,
		eventsv1beta1.SchemeBuilder,
		extensionsv1beta1.SchemeBuilder,
		flowcontrolv1beta1.SchemeBuilder,
		flowcontrolv1beta2.SchemeBuilder,
		flowcontrolv1beta3.SchemeBuilder,
		flowcontrolv1.SchemeBuilder,
		imagepolicyv1alpha1.SchemeBuilder,
		networkingv1.SchemeBuilder,
		networkingv1beta1.SchemeBuilder,
		networkingv1alpha1.SchemeBuilder,
		nodev1.SchemeBuilder,
		nodev1alpha1.SchemeBuilder,
		nodev1beta1.SchemeBuilder,
		policyv1.SchemeBuilder,
		policyv1beta1.SchemeBuilder,
		rbacv1alpha1.SchemeBuilder,
		rbacv1beta1.SchemeBuilder,
		rbacv1.SchemeBuilder,
		resourceapi.SchemeBuilder,
		schedulingv1alpha1.SchemeBuilder,
		schedulingv1beta1.SchemeBuilder,
		schedulingv1.SchemeBuilder,
		storagev1alpha1.SchemeBuilder,
		storagev1beta1.SchemeBuilder,
		storagev1.SchemeBuilder,
		svmv1alpha1.SchemeBuilder,
	}

	scheme := runtime.NewScheme()

	for _, builder := range groups {
		builder.AddToScheme(scheme)
	}

	for gvk, typ := range scheme.AllKnownTypes() {
		if gvk.Version == "__internal" {
			continue
		}

		uri := ""
		if gvk.Group != "" {
			uri += gvk.Group + "/"
		}
		uri += gvk.Version + "/" + gvk.Kind

		obj := reflect.New(typ).Elem().Interface()
		schema := toSchema(reflect.TypeOf(obj))

		distPath := filepath.Join("schemas", gvk.Group, gvk.Version, gvk.Kind+".json")
		if err := os.MkdirAll(filepath.Dir(distPath), 0755); err != nil {
			panic(err)
		}

		file, err := os.Create(distPath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		encoder := json.NewEncoder(file)
		if err := encoder.Encode(schema); err != nil {
			panic(err)
		}
		fmt.Printf("Generated schema for %s at %s\n", uri, distPath)
	}
}
