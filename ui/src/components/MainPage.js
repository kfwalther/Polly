import React from 'react'
import { toUSD } from './Helpers'
import { StockPieChart, PieChartColors } from './StockPieChart'
import StockBarChart from './StockBarChart'
import Checkbox from './Checkbox'
import PortfolioSummary from './PortfolioSummary';
import { PortfolioHoldingsTable } from './PortfolioHoldingsTable'

// Defines the Main Page of our app.
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
        // Make entire background black.
        document.body.style.backgroundColor = "black"
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

    // Helper function to map colors to our current list of stocks, sorted by market value.
    assignTickerColors() {
        // Sort the stocks/ETFs we currently own by current market value, and map them to the colors above.
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
            (s, idx) => [s.ticker, PieChartColors[idx % 31]]
        ))
    }

    // Returns the JSX to display the stock main page.
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
        // Render the stock charts and tables for the main page.
        return (
            <>
                <PortfolioSummary summaryData={this.state.portfolioSummary} />
                <br></br>
                <h3 className="header-portcomposition">Portfolio Composition</h3>
                <Checkbox
                    label="Stocks Only"
                    checked={this.state.isStocksOnlyChecked}
                    onClick={this.onStocksOnlyCheckboxClick}
                />
                {/* Put the two pie charts in a div container so they sit horizontally adjacent. */}
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
                {/* Display our current holdings in a bar chart. */}
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
                {/* Display all the stocks/ETFs in a sortable table. */}
                <PortfolioHoldingsTable
                    holdingsData={this.state.isCurrentOnlyChecked ? this.state.stockList.filter(s => (s.marketValue > 0.0)) : this.state.stockList}
                />
            </>
        )
    }

    // Render the stock main page, or a loader screen until data is retrieved from server.
    render() {
        const curState = this.state
        // TODO: Improve this loading prompt...
        return curState.stockList.length ? this.renderStockCharts() : (
            <span>LOADING STOCKS...</span>
        )
    }
}
