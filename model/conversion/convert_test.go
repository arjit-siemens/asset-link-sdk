/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: MIT
 *
 */

package conversion

import (
	"encoding/json"
	"fmt"
	generated "github.com/industrial-asset-hub/asset-link-sdk/v3/generated/iah-discovery"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConvertToDiscoveredDevice(t *testing.T) {
	device := generateDevice("Profinet", "Device")
	discoveredDevice := ConvertToDiscoveredDevice(device)
	discoveredDeviceType := fmt.Sprintf("%s/%s", baseSchemaPrefix, "Asset#@type")
	assert.Equal(t, 16, len(discoveredDevice.Identifiers))
	assert.Equal(t, "URI", discoveredDevice.Identifiers[0].Classifiers[0].GetType())
	assert.Equal(t, discoveredDeviceType, discoveredDevice.Identifiers[0].Classifiers[0].GetValue())
}

func TestConvertFromDerivedSchemaToDiscoveredDevice(t *testing.T) {
	schemaUri := "https://schema.industrial-assets.io/sat/v0.8.2"
	device := generateDevice("SatController", "Device")
	discoveredDevice := ConvertFromDerivedSchemaToDiscoveredDevice(device, schemaUri, "SatController")
	assert.Equal(t, 16, len(discoveredDevice.Identifiers))
	assert.Equal(t, "URI", discoveredDevice.Identifiers[0].Classifiers[0].GetType())
	assert.Equal(t, "https://schema.industrial-assets.io/sat/v0.8.2/SatController#@type", discoveredDevice.Identifiers[0].Classifiers[0].GetValue())
}

type DerivedDeviceInfo struct {
	model.DeviceInfo
	PasswordProtected *bool `json:"password_protected,omitempty"`
}

func TestConvertDerivedSchemaToDiscoveredDevice(t *testing.T) {
	var satDevice *DerivedDeviceInfo
	device := generateDevice("SatController", "Device")
	satDevice = &DerivedDeviceInfo{
		DeviceInfo:        *device,
		PasswordProtected: new(bool),
	}
	*satDevice.PasswordProtected = true

	discoveredDevice := ConvertFromDerivedSchemaToDiscoveredDevice(satDevice, "https://schema.industrial-assets.io/sat/v0.8.2", "SatController")
	assert.Equal(t, 17, len(discoveredDevice.Identifiers))
	passwordProtectedFound := false
	for _, identifier := range discoveredDevice.Identifiers {
		if strings.Contains(identifier.Classifiers[0].GetValue(), "password_protected") {
			passwordProtectedFound = true
			assert.Equal(t, "true", identifier.GetText())
		}
	}
	assert.True(t, passwordProtectedFound)
}

func generateDevice(typeOfAsset string, assetName string) *model.DeviceInfo {
	device := model.NewDevice(typeOfAsset, assetName)
	timestamp := model.CreateTimestamp()
	Name := "Device"
	device.Name = &Name
	product := "test-product"
	version := "1.0.0"
	vendorName := "test-vendor"
	serialNumber := "test"
	vendor := model.Organization{
		Address:        nil,
		AlternateNames: nil,
		ContactPoint:   nil,
		Id:             "",
		Name:           &vendorName,
	}
	productSerialidentifier := model.ProductSerialIdentifier{
		IdentifierType:        nil,
		IdentifierUncertainty: nil,
		ManufacturerProduct: &model.Product{
			Id:             uuid.New().String(),
			Manufacturer:   &vendor,
			Name:           &Name,
			ProductId:      &product,
			ProductVersion: &version,
		},
		SerialNumber: &serialNumber,
	}
	device.ProductInstanceIdentifier = &productSerialidentifier

	randomMacAddress := "12:12:12:12:12:12"
	identifierUncertainty := 1
	device.MacIdentifiers = append(device.MacIdentifiers, model.MacIdentifier{
		MacAddress:            &randomMacAddress,
		IdentifierUncertainty: &identifierUncertainty,
	})

	connectionPointType := model.Ipv4ConnectivityConnectionPointTypeIpv4Connectivity
	Ipv4Address := "192.168.0.1"
	Ipv4NetMask := "255.255.255.0"
	connectionPoint := "EthernetPort"
	connectionPointTypeIpv6 := model.Ipv6ConnectivityConnectionPointTypeIpv6Connectivity
	routerIpv6Address := "fd12:3456:789a::1"
	Ipv6Address := "fd12:3456:789a::1"
	conPoint := "eth0"
	relatedConnectionPoint := model.RelatedConnectionPoint{
		ConnectionPoint:    &conPoint,
		CustomRelationship: &connectionPoint,
	}
	relatedConnectionPoints := make([]model.RelatedConnectionPoint, 0)
	relatedConnectionPoints = append(relatedConnectionPoints, relatedConnectionPoint)
	Ipv4Connectivity := model.Ipv4Connectivity{
		ConnectionPointType:     &connectionPointType,
		Id:                      "1",
		InstanceAnnotations:     nil,
		Ipv4Address:             &Ipv4Address,
		NetworkMask:             &Ipv4NetMask,
		RelatedConnectionPoints: relatedConnectionPoints,
		RouterIpv4Address:       nil,
	}
	device.ConnectionPoints = append(device.ConnectionPoints, Ipv4Connectivity)
	Ipv6Connectivity := model.Ipv6Connectivity{
		ConnectionPointType:     &connectionPointTypeIpv6,
		Id:                      "2",
		InstanceAnnotations:     nil,
		Ipv6Address:             &Ipv6Address,
		RelatedConnectionPoints: nil,
		RouterIpv6Address:       &routerIpv6Address,
	}
	device.ConnectionPoints = append(device.ConnectionPoints, Ipv6Connectivity)
	ethernetType := model.EthernetPortConnectionPointTypeEthernetPort
	EthernetPort := model.EthernetPort{
		Id:                  "3",
		ConnectionPointType: &ethernetType,
		MacAddress:          &randomMacAddress,
	}
	device.ConnectionPoints = append(device.ConnectionPoints, EthernetPort)

	state := model.ManagementStateValuesUnknown
	State := model.ManagementState{
		StateTimestamp: &timestamp,
		StateValue:     &state,
	}
	device.ManagementState = State

	reachabilityStateValue := model.ReachabilityStateValuesReached
	reachabilityState := model.ReachabilityState{
		StateTimestamp: &timestamp,
		StateValue:     &reachabilityStateValue,
	}
	device.ReachabilityState = &reachabilityState
	return device
}

const fullProfinetSchemaPrefix = "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#"

func TestDeviceTransformation(t *testing.T) {
	baseSchema := "https://common-device-management.code.siemens.io/documentation/asset-modeling/base-schema/v0.7.5/"
	t.Run("TransformDevice when provided a device with identifier value of type text transforms it successfully", func(t *testing.T) {
		testDeviceForText := &generated.DiscoveredDevice{
			Identifiers: []*generated.DeviceIdentifier{{
				Value: &generated.DeviceIdentifier_Text{
					Text: "AUMA Riester GmbH & Co.KG",
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type: "URI",
					Value: fullProfinetSchemaPrefix +
						"product_instance_identifier/manufacturer_product/manufacturer/name",
				}},
			},
			},
			Timestamp: 1000010000100010,
		}
		expectedType := "URI"
		actualResult := TransformDevice(testDeviceForText, expectedType, "v0.7.5")
		expectedResult := map[string]interface{}{
			"@type": "ProfinetDevice",
			"management_state": map[string]interface{}{
				"state_value":     "unknown",
				"state_timestamp": convertTimestampToRFC339(1000010000100010),
			},
			"@context": map[string]interface{}{
				"base":      baseSchema,
				"linkml":    "https://w3id.org/linkml/",
				"lis":       "http://rds.posccaesar.org/ontology/lis14/rdl/",
				"schemaorg": "https://schema.org/",
				"skos":      "http://www.w3.org/2004/02/skos/core#",
				"@vocab":    baseSchema,
			},
			"product_instance_identifier": map[string]interface{}{
				"manufacturer_product": map[string]interface{}{
					"manufacturer": map[string]interface{}{
						"name": "AUMA Riester GmbH & Co.KG",
					},
				},
			},
		}
		assert.Equal(t, expectedResult, actualResult)
	},
	)

	t.Run("TransformDevice when provided a device with identifier value of type int64 transforms it successfully", func(t *testing.T) {
		testDeviceForInt := &generated.DiscoveredDevice{
			Identifiers: []*generated.DeviceIdentifier{{
				Value: &generated.DeviceIdentifier_Int64Value{
					Int64Value: int64(1),
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#test-1",
				}},
			},
			},
			Timestamp: 1000010000100010,
		}
		expectedType := "URI"
		actualResult := TransformDevice(testDeviceForInt, expectedType, "v0.7.5")
		expectedResult := map[string]interface{}{
			"@type": "ProfinetDevice",
			"management_state": map[string]interface{}{
				"state_value":     "unknown",
				"state_timestamp": convertTimestampToRFC339(1000010000100010),
			},
			"@context": map[string]interface{}{
				"base":      baseSchema,
				"linkml":    "https://w3id.org/linkml/",
				"lis":       "http://rds.posccaesar.org/ontology/lis14/rdl/",
				"schemaorg": "https://schema.org/",
				"skos":      "http://www.w3.org/2004/02/skos/core#",
				"@vocab":    baseSchema,
			},
			"test-1": int64(1),
		}
		assert.Equal(t, expectedResult, actualResult)
	},
	)
	t.Run("TransformDevice when provided a device with identifier value of type float64 transforms it successfully", func(t *testing.T) {
		testDeviceForFloat := &generated.DiscoveredDevice{
			Identifiers: []*generated.DeviceIdentifier{{
				Value: &generated.DeviceIdentifier_Float64Value{
					Float64Value: float64(0.1),
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#test-2/A/B",
				}},
			},
			},
			Timestamp: 1000010000100010,
		}
		expectedType := "URI"
		actualResult := TransformDevice(testDeviceForFloat, expectedType, "v0.7.5")
		expectedResult := map[string]interface{}{
			"@type": "ProfinetDevice",
			"management_state": map[string]interface{}{
				"state_value":     "unknown",
				"state_timestamp": convertTimestampToRFC339(1000010000100010),
			},
			"@context": map[string]interface{}{
				"base":      baseSchema,
				"linkml":    "https://w3id.org/linkml/",
				"lis":       "http://rds.posccaesar.org/ontology/lis14/rdl/",
				"schemaorg": "https://schema.org/",
				"skos":      "http://www.w3.org/2004/02/skos/core#",
				"@vocab":    baseSchema,
			},
			"test-2": map[string]interface{}{
				"A": map[string]interface{}{
					"B": 0.1,
				},
			},
		}
		assert.Equal(t, expectedResult, actualResult)
	},
	)
	t.Run("TransformDevice when provided a device with identifier value of type rawData transforms it successfully", func(t *testing.T) {
		testDeviceForRawData := &generated.DiscoveredDevice{
			Identifiers: []*generated.DeviceIdentifier{{
				Value: &generated.DeviceIdentifier_RawData{
					RawData: []byte{1},
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#test-2/A/B",
				}},
			},
			},
			Timestamp: 1000010000100010,
		}
		expectedType := "URI"
		actualResult := TransformDevice(testDeviceForRawData, expectedType, "v0.7.5")
		expectedResult := map[string]interface{}{
			"@type": "ProfinetDevice",
			"management_state": map[string]interface{}{
				"state_value":     "unknown",
				"state_timestamp": convertTimestampToRFC339(1000010000100010),
			},
			"@context": map[string]interface{}{
				"base":      baseSchema,
				"linkml":    "https://w3id.org/linkml/",
				"lis":       "http://rds.posccaesar.org/ontology/lis14/rdl/",
				"schemaorg": "https://schema.org/",
				"skos":      "http://www.w3.org/2004/02/skos/core#",
				"@vocab":    baseSchema,
			},
			"test-2": map[string]interface{}{
				"A": map[string]interface{}{
					"B": []byte{1},
				},
			},
		}
		assert.Equal(t, expectedResult, actualResult)
	},
	)
	t.Run("TransformDevice when provided a device with identifier value of type children transforms it successfully", func(t *testing.T) {
		testDeviceForChildren := &generated.DiscoveredDevice{
			Identifiers: []*generated.DeviceIdentifier{{
				Value: &generated.DeviceIdentifier_Children{
					Children: &generated.DeviceIdentifierValueList{
						Value: []*generated.DeviceIdentifier{
							{
								Value: &generated.DeviceIdentifier_Text{
									Text: "test-connection-point",
								},
								Classifiers: []*generated.SemanticClassifier{{
									Type:  "URI",
									Value: fullProfinetSchemaPrefix + "connection_points/related_connection_points/connection_point",
								}},
							},
						},
					},
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#connection_points",
				}},
			},
			},
			Timestamp: 1000010000100010,
		}
		expectedType := "URI"
		actualResult := TransformDevice(testDeviceForChildren, expectedType, "v0.7.5")
		expectedResult := map[string]interface{}{
			"@type": "ProfinetDevice",
			"@context": map[string]interface{}{
				"base":      baseSchema,
				"linkml":    "https://w3id.org/linkml/",
				"lis":       "http://rds.posccaesar.org/ontology/lis14/rdl/",
				"schemaorg": "https://schema.org/",
				"skos":      "http://www.w3.org/2004/02/skos/core#",
				"@vocab":    baseSchema,
			},
			"management_state": map[string]interface{}{
				"state_value":     "unknown",
				"state_timestamp": convertTimestampToRFC339(1000010000100010),
			},
			"connection_points": []map[string]interface{}{
				{"related_connection_points": map[string]interface{}{"connection_point": "test-connection-point"}}},
		}
		assert.Equal(t, expectedResult, actualResult)
	},
	)
	t.Run("TransformDevice when provided a device with several identifiers transforms it successfully", func(t *testing.T) {
		testDevice := &generated.DiscoveredDevice{
			Identifiers: []*generated.DeviceIdentifier{{
				Value: &generated.DeviceIdentifier_Children{
					Children: &generated.DeviceIdentifierValueList{
						Value: []*generated.DeviceIdentifier{
							{
								Value: &generated.DeviceIdentifier_Text{
									Text: "another-child-value",
								},
								Classifiers: []*generated.SemanticClassifier{{
									Type:  "URI",
									Value: fullProfinetSchemaPrefix + "parent-property/another-child-property",
								}},
							},
						},
					},
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#parent-property",
				}},
			},
				{
					Value: &generated.DeviceIdentifier_Children{
						Children: &generated.DeviceIdentifierValueList{
							Value: []*generated.DeviceIdentifier{
								{
									Value: &generated.DeviceIdentifier_Text{
										Text: "child-value",
									},
									Classifiers: []*generated.SemanticClassifier{{
										Type:  "URI",
										Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#parent-property/child-property",
									}},
								},
							},
						},
					},
					Classifiers: []*generated.SemanticClassifier{{
						Type:  "URI",
						Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#parent-property",
					}},
				},
				{
					Value: &generated.DeviceIdentifier_Text{
						Text: "test-serial-number",
					},
					Classifiers: []*generated.SemanticClassifier{{
						Type:  "URI",
						Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#product_instance_identifier/serial_number",
					}},
				},
			},

			Timestamp: 1000010000100010,
		}
		expectedType := "URI"
		actualResult := TransformDevice(testDevice, expectedType, "v0.7.5")
		expectedResult := map[string]interface{}{
			"@type": "ProfinetDevice",
			"@context": map[string]interface{}{
				"base":      baseSchema,
				"linkml":    "https://w3id.org/linkml/",
				"lis":       "http://rds.posccaesar.org/ontology/lis14/rdl/",
				"schemaorg": "https://schema.org/",
				"skos":      "http://www.w3.org/2004/02/skos/core#",
				"@vocab":    baseSchema,
			},
			"management_state": map[string]interface{}{
				"state_value":     "unknown",
				"state_timestamp": convertTimestampToRFC339(1000010000100010),
			},
			"parent-property": []map[string]interface{}{
				{"another-child-property": "another-child-value"},
				{"child-property": "child-value"}},
			"product_instance_identifier": map[string]interface{}{
				"serial_number": "test-serial-number",
			},
		}
		assert.Equal(t, expectedResult, actualResult)
	},
	)
}

func TestFilter(t *testing.T) {
	t.Run("filter when provided with invalid classifier type should filter out classifier", func(t *testing.T) {
		testIdentifier := &generated.DeviceIdentifier{
			Value: &generated.DeviceIdentifier_Text{
				Text: "test-text",
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "test",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#A",
			}},
		}
		result := filter(testIdentifier, "URI", fullProfinetSchemaPrefix)
		assert.Nil(t, result)
	})
	t.Run("filter when provided with valid classifier type should not filter out", func(t *testing.T) {
		testIdentifier := &generated.DeviceIdentifier{
			Value: &generated.DeviceIdentifier_Text{
				Text: "test-text",
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/ProfinetDevice#A",
			}},
		}
		result := filter(testIdentifier, "URI", fullProfinetSchemaPrefix)
		assert.NotNil(t, result)
	})
}

func TestTransformKeys(t *testing.T) {
	t.Run("transformKeys when provided with keys it assigns transformed property to device", func(t *testing.T) {
		testKeys := []string{"prop1", "prop2", "prop3", "prop4"}
		testDevice := map[string]interface{}{
			"prop1": map[string]interface{}{},
		}
		expectedDevice := map[string]interface{}{
			"prop1": map[string]interface{}{
				"prop2": map[string]interface{}{
					"prop3": map[string]interface{}{
						"prop4": "test-value",
					},
				},
			},
		}
		transformKeys(testKeys, "test-value", testDevice)
		assert.Equal(t, expectedDevice, testDevice)
	})

}

func TestRetrieveAssetTypeFromDiscoveredDeviceForValidClassifier(t *testing.T) {
	assetTypeToCheck := "ProfinetDevice"
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Text{
				Text: "test-text",
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/" + assetTypeToCheck + "#A",
			}},
		}},
	}
	assetType := retrieveAssetTypeFromDiscoveredDevice(&discoveredDevice)
	assert.Equal(t, assetTypeToCheck, assetType)
}

func TestRetrieveAssetTypeFromDiscoveredDeviceWithoutAnyIdentifier(t *testing.T) {
	assetTypeToCheck := "Asset"
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{},
	}
	assetType := retrieveAssetTypeFromDiscoveredDevice(&discoveredDevice)
	assert.Equal(t, assetTypeToCheck, assetType)
}

func TestRetrieveAssetTypeFromDiscoveredDeviceWithUnsupportedClassifier(t *testing.T) {
	assetTypeToCheck := "Asset"
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Text{
				Text: "test-text",
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "any-other-test-type-that-is-not-supported",
				Value: "unsupported-value",
			}},
		}},
	}
	assetType := retrieveAssetTypeFromDiscoveredDevice(&discoveredDevice)
	assert.Equal(t, assetTypeToCheck, assetType)
}

func TestRetrieveAssetTypeFromDiscoveredDeviceWithInvalidClassifier(t *testing.T) {
	assetTypeToCheck := "Asset"
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Text{
				Text: "test-text",
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/un-supported-uri-format",
			}},
		}},
	}
	assetType := retrieveAssetTypeFromDiscoveredDevice(&discoveredDevice)
	assert.Equal(t, assetTypeToCheck, assetType)
}

func TestMapManyArrayElementsViaMultipleArrayContainersIntoIahDevice(t *testing.T) {
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{
			{
				Value: &generated.DeviceIdentifier_Children{
					Children: &generated.DeviceIdentifierValueList{
						Value: []*generated.DeviceIdentifier{
							{
								Value: &generated.DeviceIdentifier_Text{
									Text: "element-1",
								},
								Classifiers: []*generated.SemanticClassifier{{
									Type:  "URI",
									Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array/id",
								}},
							},
							{
								Value: &generated.DeviceIdentifier_Text{
									Text: "element-1-name",
								},
								Classifiers: []*generated.SemanticClassifier{{
									Type:  "URI",
									Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array/name",
								}},
							},
						},
					},
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array",
				}},
			},
			{
				Value: &generated.DeviceIdentifier_Children{
					Children: &generated.DeviceIdentifierValueList{
						Value: []*generated.DeviceIdentifier{
							{
								Value: &generated.DeviceIdentifier_Text{
									Text: "element-2",
								},
								Classifiers: []*generated.SemanticClassifier{{
									Type:  "URI",
									Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array/id",
								}},
							},
							{
								Value: &generated.DeviceIdentifier_Text{
									Text: "element-2-name",
								},
								Classifiers: []*generated.SemanticClassifier{{
									Type:  "URI",
									Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array/name",
								}},
							},
						},
					},
				},
				Classifiers: []*generated.SemanticClassifier{{
					Type:  "URI",
					Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array",
				}},
			},
		},
	}
	iahDevice := TransformDevice(&discoveredDevice, "URI", "")
	assert.Equal(t, len(iahDevice["array"].([]map[string]interface{})), 2)
	assert.Equal(t, len(iahDevice["array"].([]map[string]interface{})[0]), 2)
	assert.Equal(t, len(iahDevice["array"].([]map[string]interface{})[1]), 2)
}

func TestMapManyArrayElementsIntoIahDevice(t *testing.T) {
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Children{
				Children: &generated.DeviceIdentifierValueList{
					Value: []*generated.DeviceIdentifier{
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B0%5D%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B0%5D%2Fname",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B1%5D%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B1%5D%2Fname",
							}},
						},
					},
				},
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array",
			}},
		}}}
	iahDevice := TransformDevice(&discoveredDevice, "URI", "")

	arrayProperty, ok := iahDevice["array"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(arrayProperty))
	assert.Equal(t, 2, len(arrayProperty[0]))
	assert.Equal(t, 2, len(arrayProperty[1]))
	assert.Equal(t, "some-name-1", arrayProperty[0]["name"])
	assert.Equal(t, "some-name-2", arrayProperty[1]["name"])
}

func TestMapManyMixedArrayElementsIntoIahDevice(t *testing.T) {
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Children{
				Children: &generated.DeviceIdentifierValueList{
					Value: []*generated.DeviceIdentifier{
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B0%5D%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B1%5D%2Fname",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B0%5D%2Fname",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array%5B1%5D%2Fid",
							}},
						},
					},
				},
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#array",
			}},
		}}}
	iahDevice := TransformDevice(&discoveredDevice, "URI", "")
	arrayProperty, ok := iahDevice["array"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(arrayProperty))
	assert.Equal(t, 2, len(arrayProperty[0]))
	assert.Equal(t, 2, len(arrayProperty[1]))
	assert.Equal(t, "some-name-1", arrayProperty[0]["name"])
	assert.Equal(t, "some-name-2", arrayProperty[1]["name"])
}

func TestMapManyDeepArrayElementsIntoIahDevice(t *testing.T) {
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Children{
				Children: &generated.DeviceIdentifierValueList{
					Value: []*generated.DeviceIdentifier{
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fname",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fname",
							}},
						},
					},
				},
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray",
			}},
		}}}
	iahDevice := TransformDevice(&discoveredDevice, "URI", "")
	arrayProperty, ok := iahDevice["a"].(map[string]interface{})["deeper"].(map[string]interface{})["array"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(arrayProperty))
	assert.Equal(t, 2, len(arrayProperty[0]))
	assert.Equal(t, 2, len(arrayProperty[1]))
	assert.Equal(t, "element-1", arrayProperty[0]["id"])
	assert.Equal(t, "element-2", arrayProperty[1]["id"])
}

func TestMapManyDeepArrayElementsWithDeepPathsIntoIahDevice(t *testing.T) {
	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Children{
				Children: &generated.DeviceIdentifierValueList{
					Value: []*generated.DeviceIdentifier{
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-1",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fname",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "element-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "some-name-2",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fname",
							}},
						},
					},
				},
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray",
			}},
		}}}
	iahDevice := TransformDevice(&discoveredDevice, "URI", "")

	arrayProperty, ok := iahDevice["a"].(map[string]interface{})["deeper"].(map[string]interface{})["array"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(arrayProperty))
	assert.Equal(t, "element-1", arrayProperty[0]["some_object"].(map[string]interface{})["id"])
	assert.Equal(t, "element-2", arrayProperty[1]["some_object"].(map[string]interface{})["id"])
}

func TestMapManyDeepArrayElementsWithDeepPathsThatContainArraysIntoIahDevice(t *testing.T) {
	expectedResult := map[string]interface{}{"@context": map[string]interface{}{"@vocab": "https://common-device-management.code.siemens.io/documentation/asset-modeling/base-schema/v0.7.5/", "base": "https://common-device-management.code.siemens.io/documentation/asset-modeling/base-schema/v0.7.5/", "linkml": "https://w3id.org/linkml/", "lis": "http://rds.posccaesar.org/ontology/lis14/rdl/", "schemaorg": "https://schema.org/", "skos": "http://www.w3.org/2004/02/skos/core#"}, "@type": "device", "a": map[string]interface{}{"deeper": map[string]interface{}{"array": []map[string]interface{}{{"some_object": map[string]interface{}{"connection_points": []map[string]interface{}{{"id": "array-0-connection-point-0", "related_connection_points": []map[string]interface{}{{"id": "array-0-con-point-0-related-connection-point-0"}, {"id": "array-0-con-point-0-related-connection-point-1"}}}, {}, {"related_connection_points": []map[string]interface{}{{}, {}, {"id": "array-0-con-point-2-related-connection-point-2"}, {"id": "array-0-con-point-2-related-connection-point-3"}}}}, "name": "array-0-name"}}, {"some_object": map[string]interface{}{"connection_points": []map[string]interface{}{{"id": "array-1-connection-point-0", "related_connection_points": []map[string]interface{}{{"id": "array-1-con-point-0-related-connection-point-0"}, {"id": "array-1-con-point-0-related-connection-point-1"}}}, {"id": "array-1-connection-point-1", "related_connection_points": []map[string]interface{}{{}, {}, {"id": "array-1-con-point-1-related-connection-point-2"}, {"id": "array-1-con-point-1-related-connection-point-3"}}}}, "id": "array-1-id", "name": "array-1-name"}}, {"some_object": map[string]interface{}{"id": "array-2-id"}}}}}, "management_state": map[string]interface{}{"state_timestamp": convertTimestampToRFC339(1000010000100010), "state_value": "unknown"}}

	discoveredDevice := generated.DiscoveredDevice{
		Identifiers: []*generated.DeviceIdentifier{{
			Value: &generated.DeviceIdentifier_Children{
				Children: &generated.DeviceIdentifierValueList{
					Value: []*generated.DeviceIdentifier{
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "array-1-id",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fid",
							}},
						},
						{
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points",
							}},
							Value: &generated.DeviceIdentifier_Children{
								Children: &generated.DeviceIdentifierValueList{
									Value: []*generated.DeviceIdentifier{
										{
											Value: &generated.DeviceIdentifier_Text{
												Text: "array-0-connection-point-0",
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Fid",
											}},
										},
										{
											Value: &generated.DeviceIdentifier_Children{
												Children: &generated.DeviceIdentifierValueList{
													Value: []*generated.DeviceIdentifier{
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-0-con-point-0-related-connection-point-0",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Frelated_connection_points%5B0%5D%2Fid",
															}},
														},
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-0-con-point-0-related-connection-point-1",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Frelated_connection_points%5B1%5D%2Fid",
															}},
														},
													},
												},
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Frelated_connection_points",
											}},
										},
										{
											Value: &generated.DeviceIdentifier_Children{
												Children: &generated.DeviceIdentifierValueList{
													Value: []*generated.DeviceIdentifier{
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-0-con-point-2-related-connection-point-2",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B2%5D%2Frelated_connection_points%5B2%5D%2Fid",
															}},
														},
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-0-con-point-2-related-connection-point-3",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B2%5D%2Frelated_connection_points%5B3%5D%2Fid",
															}},
														},
													},
												},
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fconnection_points%5B2%5D%2Frelated_connection_points",
											}},
										},
									},
								},
							},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "array-0-name",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B0%5D%2Fsome_object%2Fname",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "array-2-id",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B2%5D%2Fsome_object%2Fid",
							}},
						},
						{
							Value: &generated.DeviceIdentifier_Text{
								Text: "array-1-name",
							},
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fname",
							}},
						},
						{
							Classifiers: []*generated.SemanticClassifier{{
								Type:  "URI",
								Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points",
							}},
							Value: &generated.DeviceIdentifier_Children{
								Children: &generated.DeviceIdentifierValueList{
									Value: []*generated.DeviceIdentifier{
										{
											Value: &generated.DeviceIdentifier_Text{
												Text: "array-1-connection-point-0",
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Fid",
											}},
										},
										{
											Value: &generated.DeviceIdentifier_Children{
												Children: &generated.DeviceIdentifierValueList{
													Value: []*generated.DeviceIdentifier{
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-1-con-point-0-related-connection-point-0",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Frelated_connection_points%5B0%5D%2Fid",
															}},
														},
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-1-con-point-0-related-connection-point-1",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Frelated_connection_points%5B1%5D%2Fid",
															}},
														},
													},
												},
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B0%5D%2Frelated_connection_points",
											}},
										},
										{
											Value: &generated.DeviceIdentifier_Children{
												Children: &generated.DeviceIdentifierValueList{
													Value: []*generated.DeviceIdentifier{
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-1-con-point-1-related-connection-point-2",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B1%5D%2Frelated_connection_points%5B2%5D%2Fid",
															}},
														},
														{
															Value: &generated.DeviceIdentifier_Text{
																Text: "array-1-con-point-1-related-connection-point-3",
															},
															Classifiers: []*generated.SemanticClassifier{{
																Type:  "URI",
																Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B1%5D%2Frelated_connection_points%5B3%5D%2Fid",
															}},
														},
													},
												},
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B1%5D%2Frelated_connection_points",
											}},
										},
										{
											Value: &generated.DeviceIdentifier_Text{
												Text: "array-1-connection-point-1",
											},
											Classifiers: []*generated.SemanticClassifier{{
												Type:  "URI",
												Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray%5B1%5D%2Fsome_object%2Fconnection_points%5B1%5D%2Fid",
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			Classifiers: []*generated.SemanticClassifier{{
				Type:  "URI",
				Value: "https://schema.industrial-assets.io/profinet/1.0.0/device#a%2Fdeeper%2Farray",
			}},
		}}, Timestamp: 1000010000100010}
	iahDevice := TransformDevice(&discoveredDevice, "URI", "v0.7.5")

	assert.Equal(t, expectedResult, iahDevice)
}
func TestPanics(t *testing.T) {

	t.Run("transformKeys panics if empty key is found", func(t *testing.T) {
		testKeys := []string{}
		testDevice := map[string]interface{}{}
		assert.NotPanics(t, func() { transformKeys(testKeys, "", testDevice) })
	})

	t.Run("transformKeys panics if ip v6 connection point is provided", func(t *testing.T) {
		testKeys := []string{"prop1", "prop2", "prop3", "prop4"}

		testDevice := make(map[string]interface{})
		assert.NotPanics(t, func() { transformKeys(testKeys, "test-value", testDevice) })
	})
}

func TestCheckConversionForFile(t *testing.T) {
	t.Run("validate PNAS output against IAH schema", func(t *testing.T) {
		t.Skip("Skipping test as it is only used to get an end to end example going manually")
		testDevice := map[string]interface{}{}
		var scannedPnasResponse generated.DiscoverResponse
		// open file for reading
		file, err := os.Open("./example_dummy_device.json")
		if err != nil {
			t.Fatalf("failed to open input file: %v", err)
		}
		defer file.Close()

		// create result file for writing
		resultFile, err := os.Create("./mapped_device_after_refactor.json")
		if err != nil {
			t.Fatalf("failed to create output file: %v", err)
		}
		defer resultFile.Close()

		// get file size for creating a matching buffer
		fileInfo, _ := file.Stat()
		byteBuffer := make([]byte, fileInfo.Size())
		_, _ = file.Read(byteBuffer)
		unmarshalOptions := protojson.UnmarshalOptions{
			DiscardUnknown: true,
			AllowPartial:   true,
		}
		if err := unmarshalOptions.Unmarshal(byteBuffer, &scannedPnasResponse); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}
		// do the actual schema transformation
		testDevice = TransformDevice(scannedPnasResponse.Devices[0], "URI", "")

		// Write the result to the output file
		jsonWriter := json.NewEncoder(resultFile)
		if err := jsonWriter.Encode(testDevice); err != nil {
			t.Fatalf("failed to write result to output file: %v", err)
		}
	})
}