package service

import (
	"fmt"
	pb "high-end-lines/internal/grpc/proto/proto_go"
	"io"
	"log"
)

var _ pb.SearchServiceServer = &SearchServiceImpl{}

var orderBatchSize int = 1

type OrderServiceImpl struct {
	pb.UnimplementedOrderManagementServer
}

func (s *OrderServiceImpl) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	//time.Sleep(4 * time.Second)
	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order : %s", orderId)
		if err == io.EOF {
			log.Printf("EOF : %s", orderId)
			for _, combinedShipment := range combinedShipmentMap {
				if err := stream.Send(&combinedShipment); err != nil {
					fmt.Println("EOF Send err", err)
					return err
				}
			}
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}

		shipment, found := combinedShipmentMap[orderId.GetValue()]

		if found {
			shipment.OrderList = append(shipment.OrderList, &pb.Order{Id: orderId.GetValue()})
			combinedShipmentMap[orderId.GetValue()] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: orderId.GetValue(), Status: "Processed!"}
			comShip.OrderList = append(shipment.OrderList, &pb.Order{Id: orderId.GetValue()})
			combinedShipmentMap[orderId.GetValue()] = comShip
			log.Println("len ->", len(comShip.OrderList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrderList))
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}
