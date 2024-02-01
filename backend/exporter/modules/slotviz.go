package modules

type slotVizData struct {
	ModuleContext
}

func NewSlotVizDataModule(moduleContext ModuleContext) ModuleInterfaceSlot {
	return &slotVizData{
		ModuleContext: moduleContext,
	}
}

func (d *slotVizData) Start(slot int64) {

	//head, err := d.CL.GetChainHead()

}
