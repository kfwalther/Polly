import React from 'react'
import StockTable from './StockTable'
import StockPieChart from './StockPieChart'

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
        // Define the options for this pie chart.
        var chartOptions = {
            legend: 'none',
            pieSliceText: 'label',
            pieSliceTextStyle: { fontSize: 10 },
            sliceVisibilityThreshold: .005,
            chartArea: { top: 0, bottom: 50, left: 25, right: 25 }
        }
        // Render the stock charts and tables.
        return (
            <>
                <h3>Portfolio Composition</h3>
                <StockPieChart
                    chartData={this.state.stockList}
                    chartOptions={chartOptions}
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
