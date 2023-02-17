import React from 'react'
import { StockTable } from './StockTable'
import { toUSD } from './Helpers'
import StockPieChart from './StockPieChart'
import Checkbox from './Checkbox'
import PortfolioSummary from './PortfolioSummary';

export default class StockList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            stockList: [],
            portfolioSummary: {},
            isStocksOnlyChecked: true,
            isCurrentOnlyChecked: true,
        };
        this.serverRequest = this.serverRequest.bind(this);
        this.renderStockCharts = this.renderStockCharts.bind(this);
        this.render = this.render.bind(this);
        // Assign this instance to a global variable.
        window.stockList = this;
    }

    // Fetch the stock list from the server.
    serverRequest() {
        fetch("http://localhost:5000/securities")
            .then(response => response.json())
            .then(resp => this.setState({ stockList: resp["securities"] }))
        fetch("http://localhost:5000/summary")
            .then(response => response.json())
            .then(resp => this.setState({ portfolioSummary: resp["summary"] }))
    }

    // Runs on component mount, to grab data from the server.
    componentDidMount() {
        this.serverRequest();
    }

    // Save the new checked state of the "Stocks only" checkbox.
    onStocksOnlyCheckboxClick = checked => {
        this.setState({ isStocksOnlyChecked: checked })
    }

    // Save the new checked state of the "Show current holdings only" checkbox.
    onCurrentOnlyCheckboxClick = checked => {
        this.setState({ isCurrentOnlyChecked: checked })
    }

    // Returns the HTML to display the stock table.
    renderStockCharts() {
        // Define the options for this pie chart.
        var chartOptions = {
            legend: 'none',
            backgroundColor: 'transparent',
            pieSliceText: 'label',
            pieSliceTextStyle: { fontSize: 10 },
            pieHole: 0.25,
            sliceVisibilityThreshold: .005,
            chartArea: { top: 0, bottom: 0, left: 25, right: 25 }
        }
        // Render the stock charts and tables.
        return (
            <>
                <PortfolioSummary summaryData={this.state.portfolioSummary}/>
                <br></br>
                <h3 className="header-portcomposition">Portfolio Composition</h3>
                <Checkbox
                    label="Stocks Only"
                    checked={this.state.isStocksOnlyChecked}
                    onClick={this.onStocksOnlyCheckboxClick} />
                <div className="charts-container">
                    <div className="chart-marketvalue">
                        <StockPieChart
                            chartData={this.state.stockList}
                            chartOptions={chartOptions}
                            displayDataset="marketValue"
                            filterOptions={this.state.isStocksOnlyChecked}
                            title={toUSD(this.state.portfolioSummary.totalMarketValue)}
                            titleDesc={"Market Value"}
                        />
                    </div>
                    <div className="chart-costbasis">
                        <StockPieChart
                            chartData={this.state.stockList}
                            chartOptions={chartOptions}
                            displayDataset="totalCostBasis"
                            filterOptions={this.state.isStocksOnlyChecked}
                            title={toUSD(this.state.portfolioSummary.totalCostBasis)}
                            titleDesc={"Cost Basis"}
                        />
                    </div>
                </div>

                <Checkbox
                    label="Show Current Holdings Only"
                    checked={this.state.isCurrentOnlyChecked}
                    onClick={this.onCurrentOnlyCheckboxClick} />
                <StockTable
                    data={this.state.isCurrentOnlyChecked ? this.state.stockList.filter(s => (s.marketValue > 0.0)) : this.state.stockList}
                />
            </>
        )
    }

    // Render the stock table, or a loader screen until data is retrieved from server.
    render() {
        const curState = this.state
        return curState.stockList.length ? this.renderStockCharts() : (
            <span>LOADING STOCKS...</span>
        )
    }
}
