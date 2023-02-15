import React from 'react'
import StockTable from './StockTable'
import { Chart } from 'react-google-charts';

export default class StockList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            stockList: [],
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
            .then(secs => this.setState({ stockList: secs["securities"] }))
    }

    // Runs on component mount, to grab data from the server.
    componentDidMount() {
        this.serverRequest();
    }

    // Returns the HTML to display the stock table.
    renderStockCharts() {
        // TODO: Put all the pie chart stuff in a separate file.
        let filtered = this.state.stockList.filter(a => (a.marketValue >= 0.0 && a.securityType == "Stock"));
        // Sort the stocks by current market value.
        let sorted = filtered.sort((a, b) => b.marketValue - a.marketValue);
        // Put the sorted values in a map, and add a column header.
        let pieChartData = sorted.map(x => [x.ticker, x.marketValue])
        pieChartData.unshift(['Ticker', 'Market Value'])
        // Return the charted stock data.
        var chartOptions = {
            legend: 'none',
            pieSliceText: 'label',
            pieSliceTextStyle: { fontSize: 10 },
            sliceVisibilityThreshold: .005,
            chartArea: { top: 0, bottom: 50 }
        }

        return (
            <>
                <h3>Portfolio Composition</h3>
                <Chart
                    chartType="PieChart"
                    data={pieChartData}
                    options={chartOptions}
                    width={"100%"}
                    height={"750px"}
                />
                <StockTable data={this.state.stockList} />
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
