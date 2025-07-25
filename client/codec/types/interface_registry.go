// Sourced with modifications from https://github.com/cosmos/cosmos-sdk/blob/d1b5b0c5ae2c51206cc1849e09e4d59986742cc3/codec/types/interface_registry.go
package types

import (
	"fmt"
	"reflect"

	cosmosCodecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
)

type ProbeInterfaceRegistry struct {
	cosmosCodecTypes.InterfaceRegistry

	interfaceNames map[string]reflect.Type
	interfaceImpls map[reflect.Type]interfaceMap
	implInterfaces map[reflect.Type]reflect.Type
	typeURLMap     map[string]reflect.Type
}

type interfaceMap = map[string]reflect.Type

// NewInterfaceRegistry returns a new InterfaceRegistry
func NewInterfaceRegistry() (cosmosCodecTypes.InterfaceRegistry, *ProbeInterfaceRegistry) {

	probeRegistry := &ProbeInterfaceRegistry{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
		implInterfaces: map[reflect.Type]reflect.Type{},
		typeURLMap:     map[string]reflect.Type{},
	}

	return probeRegistry, probeRegistry
}

func (registry *ProbeInterfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...proto.Message) {
	typ := reflect.TypeOf(iface)
	if typ.Elem().Kind() != reflect.Interface {
		panic(fmt.Errorf("%T is not an interface type", iface))
	}

	registry.interfaceNames[protoName] = typ
	registry.RegisterImplementations(iface, impls...)
}

// EnsureRegistered ensures there is a registered interface for the given concrete type.
//
// Returns an error if not, and nil if so.
func (registry *ProbeInterfaceRegistry) EnsureRegistered(impl interface{}) error {
	if reflect.ValueOf(impl).Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", impl)
	}

	if _, found := registry.implInterfaces[reflect.TypeOf(impl)]; !found {
		return fmt.Errorf("%T does not have a registered interface", impl)
	}

	return nil
}

// RegisterImplementations registers a concrete proto Message which implements
// the given interface.
//
// This function PANICs if different concrete types are registered under the
// same typeURL.
func (registry *ProbeInterfaceRegistry) RegisterImplementations(iface interface{}, impls ...proto.Message) {
	for _, impl := range impls {
		typeURL := "/" + proto.MessageName(impl)
		registry.registerImpl(iface, typeURL, impl)
	}
}

// RegisterCustomTypeURL registers a concrete type which implements the given
// interface under `typeURL`.
//
// This function PANICs if different concrete types are registered under the
// same typeURL.
func (registry *ProbeInterfaceRegistry) RegisterCustomTypeURL(iface interface{}, typeURL string, impl proto.Message) {
	registry.registerImpl(iface, typeURL, impl)
}

func (registry *ProbeInterfaceRegistry) TypeURLIsRegistered(typeURL string) bool {
	_, found := registry.typeURLMap[typeURL]
	return found
}

// registerImpl registers a concrete type which implements the given
// interface under `typeURL`.
//
// This function PANICs if different concrete types are registered under the
// same typeURL.
func (registry *ProbeInterfaceRegistry) registerImpl(iface interface{}, typeURL string, impl proto.Message) {
	ityp := reflect.TypeOf(iface).Elem()
	imap, found := registry.interfaceImpls[ityp]
	if !found {
		imap = map[string]reflect.Type{}
	}

	implType := reflect.TypeOf(impl)
	if !implType.AssignableTo(ityp) {
		panic(fmt.Errorf("type %T doesn't actually implement interface %+v", impl, ityp))
	}

	// Check if we already registered something under the given typeURL. It's
	// okay to register the same concrete type again, but if we are registering
	// a new concrete type under the same typeURL, then we throw an error (here,
	// we panic).
	foundImplType, found := imap[typeURL]
	if found && foundImplType != implType {
		panic(
			fmt.Errorf(
				"concrete type %s has already been registered under typeURL %s, cannot register %s under same typeURL. "+
					"This usually means that there are conflicting modules registering different concrete types "+
					"for a same interface implementation",
				foundImplType,
				typeURL,
				implType,
			),
		)
	}

	imap[typeURL] = implType
	registry.typeURLMap[typeURL] = implType
	registry.implInterfaces[implType] = ityp
	registry.interfaceImpls[ityp] = imap
}

func (registry *ProbeInterfaceRegistry) ListAllInterfaces() []string {
	interfaceNames := registry.interfaceNames
	keys := make([]string, 0, len(interfaceNames))
	for key := range interfaceNames {
		keys = append(keys, key)
	}
	return keys
}

func (registry *ProbeInterfaceRegistry) ListImplementations(ifaceName string) []string {
	typ, ok := registry.interfaceNames[ifaceName]
	if !ok {
		return []string{}
	}

	impls, ok := registry.interfaceImpls[typ.Elem()]
	if !ok {
		return []string{}
	}

	keys := make([]string, 0, len(impls))
	for key := range impls {
		keys = append(keys, key)
	}
	return keys
}

func (registry *ProbeInterfaceRegistry) UnpackAny(any *cosmosCodecTypes.Any, iface interface{}) error {
	// here we gracefully handle the case in which `any` itself is `nil`, which may occur in message decoding
	if any == nil {
		return nil
	}

	if any.TypeUrl == "" {
		// if TypeUrl is empty return nil because without it we can't actually unpack anything
		return nil
	}

	rv := reflect.ValueOf(iface)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("UnpackAny expects a pointer")
	}

	rt := rv.Elem().Type()

	cachedValue := any.GetCachedValue()
	if cachedValue != nil {
		if reflect.TypeOf(cachedValue).AssignableTo(rt) {
			rv.Elem().Set(reflect.ValueOf(cachedValue))
			return nil
		}
	}

	imap, found := registry.interfaceImpls[rt]
	if !found {
		return fmt.Errorf("no registered implementations of type %+v", rt)
	}

	typ, found := imap[any.TypeUrl]
	if !found {
		return fmt.Errorf("no concrete type registered for type URL %s against interface %T", any.TypeUrl, iface)
	}

	msg, ok := reflect.New(typ.Elem()).Interface().(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto unmarshal %T", msg)
	}

	err := proto.Unmarshal(any.Value, msg)
	if err != nil {
		return err
	}

	err = UnpackInterfaces(msg, registry)
	if err != nil {
		return err
	}

	rv.Elem().Set(reflect.ValueOf(msg))

	newAny, err := cosmosCodecTypes.NewAnyWithValue(msg)

	if err != nil {
		return err
	}

	var typeURL string
	// WARN: Is this the proper way to handle this? The custom message implementations are not returning a TypeURL in the Any after calling proto.MessageName
	if newAny.TypeUrl == "/" {
		typeURL = any.TypeUrl
	} else {
		typeURL = newAny.TypeUrl
	}

	*any = *newAny
	any.TypeUrl = typeURL

	return nil
}

// Resolve returns the proto message given its typeURL. It works with types
// registered with RegisterInterface/RegisterImplementations, as well as those
// registered with RegisterWithCustomTypeURL.
func (registry *ProbeInterfaceRegistry) Resolve(typeURL string) (proto.Message, error) {
	typ, found := registry.typeURLMap[typeURL]
	if !found {
		return nil, fmt.Errorf("unable to resolve type URL %s", typeURL)
	}

	msg, ok := reflect.New(typ.Elem()).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't resolve type URL %s", typeURL)
	}

	return msg, nil
}

// UnpackInterfaces is a convenience function that calls UnpackInterfaces
// on x if x implements UnpackInterfacesMessage
func UnpackInterfaces(x interface{}, unpacker cosmosCodecTypes.AnyUnpacker) error {
	if msg, ok := x.(cosmosCodecTypes.UnpackInterfacesMessage); ok {
		return msg.UnpackInterfaces(unpacker)
	}
	return nil
}
