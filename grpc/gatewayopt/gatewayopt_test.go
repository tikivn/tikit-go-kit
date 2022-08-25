package gatewayopt

import (
	"strings"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	Request struct {
		// Required
		RequestID string `protobuf:"bytes,3,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
		// Required
		Amount string `protobuf:"bytes,4,opt,name=amount,proto3" json:"amount,omitempty"`
		// Required
		OrderID string `protobuf:"bytes,5,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
		// Required
		OrderInfo string `protobuf:"bytes,6,opt,name=order_info,json=orderInfo,proto3" json:"order_info,omitempty"`
		// Required
		// Fixed = momo_wallet
		OrderType string `protobuf:"bytes,7,opt,name=order_type,json=orderType,proto3" json:"order_type,omitempty"`
		// Required
		TransID string `protobuf:"bytes,8,opt,name=trans_id,json=transId,proto3" json:"trans_id,omitempty"`
		// Required
		Message string `protobuf:"bytes,9,opt,name=message,proto3" json:"message,omitempty"`
		// Required
		ErrorCode int32 `protobuf:"varint,12,opt,name=error_code,json=errorCode,proto3" json:"error_code,omitempty"`
	}
)

func (r *Request) Reset()       {}
func (*Request) String() string { return "Request" }
func (*Request) ProtoMessage()  {}

func TestFormMarshaler_NewDecoder(t *testing.T) {
	// Arrange
	m := &formMarshaler{
		JSONPb: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true},
		},
	}

	data := "errorCode=0&requestId=978fb5c4-88ff-42bc-8b69-522eda5c56a3&amount=300000&orderId=QD122IWPGJ51&orderInfo=QD122IWPGJ51&orderType=momo_wallet&transId=2319925829&message=Success"
	decoder := m.NewDecoder(strings.NewReader(data))
	actual := &Request{}
	expected := &Request{
		RequestID: "978fb5c4-88ff-42bc-8b69-522eda5c56a3",
		Amount:    "300000",
		OrderID:   "QD122IWPGJ51",
		OrderInfo: "QD122IWPGJ51",
		OrderType: "momo_wallet",
		TransID:   "2319925829",
		Message:   "Success",
	}
	// Act
	err := decoder.Decode(actual)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
