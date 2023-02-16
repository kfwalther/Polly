import React from 'react'
import StockTable from './StockTable'
import StockPieChart from './StockPieChart'
import Checkbox from './Checkbox'

export default class StockList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            stockList: [],
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
            .then(secs => this.setState({ stockList: secs["securities"] }))
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
            pieSliceText: 'label',
            pieSliceTextStyle: { fontSize: 10 },
            pieHole: 0.3,
            sliceVisibilityThreshold: .005,
            chartArea: { top: 0, bottom: 0, left: 0, right: 0 }
        }
        // Render the stock charts and tables.
        return (
            <>
                <h3>Portfolio Composition</h3>
                <Checkbox
                    label="Stocks Only"
                    checked={this.state.isStocksOnlyChecked}
                    onClick={this.onStocksOnlyCheckboxClick} />
                <StockPieChart
                    chartData={this.state.stockList}
                    chartOptions={chartOptions}
                    filterOptions={this.state.isStocksOnlyChecked}
                />
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
