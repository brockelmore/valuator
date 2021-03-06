package valuator

import (
	"encoding/json"
	"log"
)

type Measures interface {
	Filing() Filing
	FiledOn() string
	NewYoy(Measures) error
	Yoy() Yoy
	BookValue() float64
	ContribMargin() float64
	OpsMargin() float64
	OperatingLeverage() float64
	FinancialLeverage() float64
	ReturnOnEquity() float64
	ReturnOnAssets() float64
	DividendPerShare() float64
	FreeCashFlow() float64
	PayOutToFcf() float64
	WorkingCapital() float64
	CurrentRatio() float64
	String() string
}

type measures struct {
	filing     Filing
	Date       string  `json:"Date"`
	Bv         float64 `json:"Book Value"`
	Cm         float64 `json:"Contribution Margin"`
	Om         float64 `json:"Operating Margin"`
	Ol         float64 `json:"Operating Leverage"`
	Fl         float64 `json:"Financial Leverage (%)"`
	RoE        float64 `json:"Return on Equity (%)"`
	RoA        float64 `json:"Return on Assets"`
	Div        float64 `json:"Dividend"`
	FcF        float64 `json:"Free Cash Flow"`
	DivToFcf   float64 `json:"Dividend to FCF"`
	Wc         float64 `json:"Working Capital"`
	Cr         float64 `json:"Current Ratio"`
	YearOnYear *yoy    `json:"YoY"`
}

func (m measures) String() string {
	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		log.Println(m)
		log.Fatal("Error marshaling financial data: ", err)
	}
	return string(data)
}

func NewMeasures(filing Filing, q bool) Measures {
	m := new(measures)
	m.filing = filing
	m.Date = filing.FiledOn()
	if q == false {
		m.collect()
	}
	return m
}

func (m *measures) Yoy() Yoy {
	return m.YearOnYear
}

func (m *measures) NewYoy(past Measures) error {
	yoy, err := NewYoy(past, m)
	if err == nil {
		m.YearOnYear = yoy
	}
	return err
}

func (m *measures) collect() {
	m.Bv = m.BookValue()
	m.Cm = m.ContribMargin()
	m.Om = m.OpsMargin()
	m.Ol = m.OperatingLeverage()
	m.Fl = m.FinancialLeverage()
	m.Div = m.DividendPerShare()
	m.DivToFcf = m.PayOutToFcf()
	m.FcF = m.FreeCashFlow()
	m.RoA = m.ReturnOnAssets()
	m.RoE = m.ReturnOnEquity()
	m.Wc = m.WorkingCapital()
	m.Cr = m.CurrentRatio()
}

func (m *measures) FiledOn() string {
	return m.filing.FiledOn()
}

func (m *measures) Filing() Filing {
	return m.filing
}

/*
 BookValue:
    Value of the company retained within the equity portion of the BS
		BV = TotalEquity on balance sheet/Total share count

*/
func (m *measures) BookValue() float64 {
	eq, err := m.filing.TotalEquity()
	if err != nil {
		return 0
	}
	sc, err := m.filing.ShareCount()
	if err != nil {
		return 0
	}
	return round(eq / sc)

}

func (m *measures) ContribMargin() float64 {
	rev, err := m.filing.Revenue()
	if err != nil {
		return 0
	}

	cr, err := m.filing.CostOfRevenue()
	if err != nil {
		return 0
	}

	return percentage((float64(rev) - float64(cr)) / float64(rev))
}

func (m *measures) OpsMargin() float64 {
	oi, err := m.filing.OperatingIncome()
	if err != nil {
		return 0
	}

	rev, err := m.filing.Revenue()
	if err != nil {
		return 0
	}

	return percentage(float64(oi) / float64(rev))
}

/*
 Operating leverage:
     ratio of contribution margin to operating margin
 The ratio captures the relation between material cost of revenue vs the
 running cost of revenue
 contribution margin (CM) = Margin of profit against materials cost
 Operating margin (OM) = Margin of operating income against revenue
 Operating leverage = CM/OM
*/
func (m *measures) OperatingLeverage() float64 {
	if m.OpsMargin() != 0 {
		return round(m.ContribMargin() / m.OpsMargin())
	}
	return 0
}

func (m *measures) FinancialLeverage() float64 {
	eq, err := m.filing.TotalEquity()
	if err != nil {
		return 0
	}
	ld, _ := m.filing.LongTermDebt()
	sd, _ := m.filing.ShortTermDebt()

	return percentage((ld + sd) / eq)
}

func (m *measures) ReturnOnEquity() float64 {
	ni, err := m.filing.NetIncome()
	if err != nil {
		return 0
	}
	eq, err := m.filing.TotalEquity()
	if err != nil {
		return 0
	}
	return percentage(ni / eq)
}

func (m *measures) ReturnOnAssets() float64 {
	return 0
}

func (m *measures) DividendPerShare() float64 {
	if dps, err := m.Filing().DividendPerShare(); err == nil {
		return dps
	}
	return 0
}

func (m *measures) FreeCashFlow() float64 {
	fcf, err := m.filing.OperatingCashFlow()
	if err != nil {
		return 0
	}
	return fcf
}

func (m *measures) PayOutToFcf() float64 {
	div, err := m.filing.Dividend()
	if err != nil {
		return 0
	}
	return percentage(div / m.FreeCashFlow())
}

func (m *measures) WorkingCapital() float64 {
	assets, err := m.filing.CurrentAssets()
	if err != nil {
		return 0
	}
	liab, err := m.filing.CurrentLiabilities()
	if err != nil {
		return 0
	}
	return (assets - liab)
}

func (m *measures) CurrentRatio() float64 {
	assets, err := m.filing.CurrentAssets()
	if err != nil {
		return 0
	}
	liab, err := m.filing.CurrentLiabilities()
	if err != nil {
		return 0
	}
	return round(assets / liab)
}
