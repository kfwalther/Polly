import React from 'react'
import { StockTable } from './StockTable'
import { toUSD } from './Helpers'
import StockPieChart from './StockPieChart'
import StockBarChart from './StockBarChart'
import Checkbox from './Checkbox'
import PortfolioSummary from './PortfolioSummary';

export default class MainPage extends React.Component {
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
        console.log('Refreshing data...')
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
        var pieChartOptions = {
            legend: 'none',
            backgroundColor: 'transparent',
            pieSliceText: 'label',
            pieSliceTextStyle: { fontSize: 10 },
            pieHole: 0.25,
            sliceVisibilityThreshold: .005,
            chartArea: { top: 0, bottom: 0, left: 25, right: 25 }
        }
        // Define the options for this pie chart.
        var barChartOptions = {
            backgroundColor: 'black',
            legend: { position: 'none' },
            chartArea: { top: 25, bottom: 50, left: 25, right: 25 },
            vAxis: { format: 'short', textStyle: { fontSize: 12, bold: true, color: 'grey' } },
            hAxis: { showTextEvery: 1, maxAlternation: 1, slantedText: true, slantedTextAngle: 45, textStyle: { fontSize: 12, bold: true, color: 'grey' } },
            bar: { groupWidth: '40%' }
        }
        // Render the stock charts and tables.
        return (
            <>
                <button >Refresh</button>
                <PortfolioSummary summaryData={this.state.portfolioSummary} />
                <br></br>
                <h3 className="header-portcomposition">Portfolio Composition</h3>
                <Checkbox
                    label="Stocks Only"
                    checked={this.state.isStocksOnlyChecked}
                    onClick={this.onStocksOnlyCheckboxClick}
                />
                <div className="charts-container">
                    <div className="chart-marketvalue">
                        <StockPieChart
                            chartData={this.state.stockList}
                            chartOptions={pieChartOptions}
                            displayDataset="marketValue"
                            filterOptions={this.state.isStocksOnlyChecked}
                            title={toUSD(this.state.portfolioSummary.totalMarketValue)}
                            titleDesc={"Market Value"}
                        />
                    </div>
                    <div className="chart-costbasis">
                        <StockPieChart
                            chartData={this.state.stockList}
                            chartOptions={pieChartOptions}
                            displayDataset="totalCostBasis"
                            filterOptions={this.state.isStocksOnlyChecked}
                            title={toUSD(this.state.portfolioSummary.totalCostBasis)}
                            titleDesc={"Cost Basis"}
                        />
                    </div>
                </div>
                <h3 className="header-myholdings">My Holdings</h3>
                <StockBarChart
                    chartData={this.state.stockList}
                    chartOptions={barChartOptions}
                />
                <Checkbox
                    label="Show Current Holdings Only"
                    checked={this.state.isCurrentOnlyChecked}
                    onClick={this.onCurrentOnlyCheckboxClick}
                    marginLeft="10px"
                />
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
