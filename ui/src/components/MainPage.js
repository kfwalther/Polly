import React from 'react'
import { toUSD } from './Helpers'
import { StockPieChart, PieChartColors } from './StockPieChart'
import StockBarChart from './StockBarChart'
import Checkbox from './Checkbox'
import { Button } from '@mui/material';
import Select from 'react-select';
import PortfolioSummary from './PortfolioSummary';
import { PortfolioHoldingsTable } from './PortfolioHoldingsTable'
import { PortfolioMapChart, PortfolioMapSizeSelectOptions, PortfolioMapColorSelectOptions} from './PortfolioMapChart'

// Defines the Main Page of our app.
export default class MainPage extends React.Component {
    // The MainPage constructor.
    constructor(props) {
        super(props);
        this.state = {
            stockList: [],
            fullPortfolioSummary: {},
            stockPortfolioSummary: {},
            isStocksOnlyChecked: true,
            isIncludeCashBalanceChecked: true,
            isCurrentOnlyChecked: true,
            portfolioMapSizeSelection: 'marketValue',
            portfolioMapColorSelection: 'revenueGrowthPercentageYoy',
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
        fetch("http://" + process.env.REACT_APP_API_BASE_URL + "/securities")
            .then(response => response.json())
            .then(resp => this.setState({ stockList: resp["securities"] }))
        fetch("http://" + process.env.REACT_APP_API_BASE_URL + "/summary")
            .then(response => response.json())
            .then(resp => {
                this.setState({ fullPortfolioSummary: resp["summary"][0] })
                this.setState({ stockPortfolioSummary: resp["summary"][1] })
            })
    }

    // Runs on component mount, to grab data from the server.
    componentDidMount() {
        // Make entire background black.
        document.body.style.backgroundColor = "black"
        this.serverRequest();
    }

    
    buttonClick = () => {
        // Copy the top-25 to clipboard.
        const listToExport = this.state.stockList
        .filter(s => (parseFloat(s.marketValue) > 0.0 && s.securityType == 'Stock'))
        .sort((a, b) => b.marketValue - a.marketValue)
        .slice(0, 25)
        .map(s => '$' + s.ticker)
        .join(' ')
        window.prompt('Copy to clipboard: Ctrl+C, Enter', listToExport)
    }
    
    // Save the new checked state of the checkboxes.
    onStocksOnlyCheckboxClick = checked => {
        this.setState({ isStocksOnlyChecked: checked })
    }
    onIncludeCashBalanceCheckboxClick = checked => {
        this.setState({ isIncludeCashBalanceChecked: checked })
    }
    onCurrentOnlyCheckboxClick = checked => {
        this.setState({ isCurrentOnlyChecked: checked })
    }

    refreshPortfolioMapSize = selection => {
        this.setState({ portfolioMapSizeSelection: selection.value })
    }

    refreshPortfolioMapColor = selection => {
        this.setState({ portfolioMapColorSelection: selection.value })
    }

    // Helper function to map colors to our current list of stocks, sorted by market value.
    assignTickerColors() {
        // Sort the stocks/ETFs we currently own by current market value, and map them to the colors above.
        this.tickerMap = new Map(
            // Filter out securities we no longer own.
            this.state.stockList.filter(s => (parseFloat(s.marketValue) > 0.0)
            ).sort(
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
            chartArea: { top: 25, bottom: 50, left: 40, right: 40 },
            vAxis: { format: 'short', textStyle: { fontSize: 12, bold: true, color: 'grey' } },
            hAxis: { showTextEvery: 1, maxAlternation: 1, slantedText: true, slantedTextAngle: 45, textStyle: { fontSize: 12, bold: true, color: 'grey' } },
            bar: { groupWidth: '40%' }
        }
        // Calculate some displayed values based on the current checkbox config.
        var cashBalance = this.state.stockList.find(s => s.ticker === 'CASH').marketValue
        var marketValuePieChart = toUSD(this.state.isStocksOnlyChecked ? 
                this.state.isIncludeCashBalanceChecked ? 
                        this.state.stockPortfolioSummary.totalMarketValue :
                        this.state.stockPortfolioSummary.totalMarketValue - cashBalance
                    : 
                this.state.isIncludeCashBalanceChecked ? 
                    this.state.fullPortfolioSummary.totalMarketValue :
                    this.state.fullPortfolioSummary.totalMarketValue - cashBalance)
        var costBasisPieChart = toUSD(this.state.isStocksOnlyChecked ? 
                this.state.stockPortfolioSummary.totalCostBasis :
                this.state.fullPortfolioSummary.totalCostBasis)
        // Render the stock charts and tables for the main page.
        return (
            <>
                <PortfolioSummary summaryData={this.state.fullPortfolioSummary} />
                <br></br>
                <h3 className="header-centered">Portfolio Composition</h3>
                {/* Put the two pie charts in a div container so they sit horizontally adjacent. */}
                <div className="charts-container">
                    <div className="chart-marketvalue">
                        <StockPieChart
                            chartData={this.state.stockList}
                            displayDataset="marketValue"
                            stocksOnly={this.state.isStocksOnlyChecked}
                            includeCash={this.state.isIncludeCashBalanceChecked}
                            title={marketValuePieChart}
                            titleDesc={"Market Value"}
                            tickerColors={this.tickerMap}
                        />
                    </div>
                    <div className="chart-costbasis">
                        <StockPieChart
                            chartData={this.state.stockList}
                            displayDataset="totalCostBasis"
                            stocksOnly={this.state.isStocksOnlyChecked}
                            includeCash={this.state.isIncludeCashBalanceChecked}
                            title={costBasisPieChart}
                            titleDesc={"Cost Basis"}
                            tickerColors={this.tickerMap}
                        />
                    </div>
                </div>
                <div className="checkbox-horiz-container">
                    <div className="config-checkbox-div">
                        <Checkbox
                            label="Stocks Only"
                            checked={this.state.isStocksOnlyChecked}
                            onClick={this.onStocksOnlyCheckboxClick}
                            marginLeftVal="20px"
                        />
                    </div>
                    <div className="config-checkbox-div">
                        <Checkbox
                            label="Display Cash Balance"
                            checked={this.state.isIncludeCashBalanceChecked}
                            onClick={this.onIncludeCashBalanceCheckboxClick}
                            marginLeftVal="20px"
                        />
                    </div>
                    <div className="config-checkbox-div">
                        <Checkbox
                            label="Current Holdings Only"
                            checked={this.state.isCurrentOnlyChecked}
                            onClick={this.onCurrentOnlyCheckboxClick}
                            marginLeftVal="20px"
                        />
                    </div>
                    <div className="config-checkbox-div">
                        {/* Display a button to export the top 25 stock names. */}
                        <Button 
                            className="export-button" 
                            variant="contained" 
                            color="info"
                            onClick={this.buttonClick}
                        >
                            {'Export Top 25 Stocks'}
                        </Button>
                    </div>
                </div>
                <h3 className="header-left">{'My Holdings (' + this.state.stockPortfolioSummary.totalSecurities + ' Stocks)'}</h3>
                {/* Display our current holdings in a bar chart. */}
                <StockBarChart
                    chartData={this.state.stockList}
                    chartOptions={barChartOptions}
                />
                <div className="portfoliomap-picker-container">
                    <h4 className='portfoliomap-size-picker-label'>Size by: </h4>
                    <Select 
                        options={PortfolioMapSizeSelectOptions} 
                        onChange={this.refreshPortfolioMapSize}
                        defaultValue={PortfolioMapSizeSelectOptions.filter(o => o.label === 'Market Value')} 
                    />
                    <h4 className='portfoliomap-color-picker-label'>Color by: </h4>
                    <Select 
                        options={PortfolioMapColorSelectOptions} 
                        onChange={this.refreshPortfolioMapColor}
                        defaultValue={PortfolioMapColorSelectOptions.filter(o => o.label === 'Growth Rate TTM')}
                    />
                </div>
                {/* Display our current holdings in a portfolio map chart also. */}
                <PortfolioMapChart
                    chartData={this.state.stockList}
                    sizeBy={this.state.portfolioMapSizeSelection}
                    colorBy={this.state.portfolioMapColorSelection}
                />
                {/* Display all the stocks/ETFs in a sortable table, account for user filtering selections. */}
                <PortfolioHoldingsTable
                    holdingsData={this.state.isCurrentOnlyChecked ?
                        this.state.isStocksOnlyChecked ?
                            this.state.stockList.filter(s => s.currentlyHeld && s.securityType === "Stock") :
                            this.state.stockList.filter(s => s.currentlyHeld) :
                        this.state.isStocksOnlyChecked ?
                            this.state.stockList.filter(s => s.securityType === "Stock") :
                            this.state.stockList
                    }
                    totalPortfolioValue={this.state.isStocksOnlyChecked ? 
                        this.state.stockPortfolioSummary.totalMarketValue : 
                        this.state.fullPortfolioSummary.totalMarketValue
                    }
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
