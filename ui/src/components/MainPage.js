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
        this.tickerMap = {}
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

    assignTickerColors() {
        var pieColors = [
            '#3366cc',
            '#dc3912',
            '#ff9900',
            '#109618',
            '#990099',
            '#0099c6',
            '#dd4477',
            '#66aa00',
            '#b82e2e',
            '#316395',
            '#994499',
            '#22aa99',
            '#aaaa11',
            '#6633cc',
            '#e67300',
            '#8b0707',
            '#651067',
            '#329262',
            '#5574a6',
            '#3b3eac',
            '#b77322',
            '#16d620',
            '#b91383',
            '#f4359e',
            '#9c5935',
            '#a9c413',
            '#2a778d',
            '#668d1c',
            '#bea413',
            '#0c5922',
            '#743411'
        ]
        this.tickerMap = new Map(this.state.stockList.filter(s => {
            // Filter out securities we no longer own.
            if (parseFloat(s.marketValue) > 0.0) {
                return s
            }
        }).sort(
            // Sort the remaining by current value
            (a, b) => b.marketValue - a.marketValue
        ).map(
            // Map the sorted tickers to colors (rolling over after 31).
            (s, idx) => [s.ticker, pieColors[idx % 31]]
        ))
    }

    // Returns the HTML to display the stock table.
    renderStockCharts() {
        // Map ticker names to pie chart colors.
        this.assignTickerColors()
        // Define the options for the bar chart.
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
                            displayDataset="marketValue"
                            filterOptions={this.state.isStocksOnlyChecked}
                            title={toUSD(this.state.portfolioSummary.totalMarketValue)}
                            titleDesc={"Market Value"}
                            tickerColors={this.tickerMap}
                        />
                    </div>
                    <div className="chart-costbasis">
                        <StockPieChart
                            chartData={this.state.stockList}
                            displayDataset="totalCostBasis"
                            filterOptions={this.state.isStocksOnlyChecked}
                            title={toUSD(this.state.portfolioSummary.totalCostBasis)}
                            titleDesc={"Cost Basis"}
                            tickerColors={this.tickerMap}
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
