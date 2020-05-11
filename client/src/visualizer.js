import React, { Component, createRef } from "react";
import PropTypes from "prop-types";

class Visualizer extends Component {
  constructor(props) {
    super(props);
    this.canvas = createRef();
    this.container = createRef();
    this.state = { width: 0, height: 0 };
  }

  componentDidMount() {
    const { stream } = this.props;

    // Resize canvas to match container.
    const { canvas, container } = this;
    {
      const { clientWidth, clientHeight } = container.current;
      this.setState({ width: clientWidth, height: clientHeight });
    }

    // Create audio and canvas contexts.
    this.ctx = canvas.current.getContext("2d");
    const acx = new AudioContext();

    // Create analyzer from audio context.
    const analyzer = acx.createAnalyser();
    analyzer.fftSize = 2048;
    analyzer.smoothingTimeConstant = 0.6;
    this.frequency = new Float32Array(analyzer.frequencyBinCount);

    // Connect source to analyzer.
    this.source = acx.createMediaStreamSource(stream);
    this.source.connect(analyzer);
    this.analyzer = analyzer;

    // Begin update loop.
    this.update();
  }

  componentWillUnmount() {
    this.source.disconnect(this.analyzer);
  }

  update = () => {
    const { color } = this.props;
    const { ctx, analyzer, frequency } = this;
    const { current: canvas } = this.canvas;
    if (!canvas) return;

    // Request next animation frame.
    requestAnimationFrame(this.update, canvas);

    // Reset canvas.
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = color;
    analyzer.getFloatFrequencyData(frequency);

    const barWidth = 5;
    const barSpace = 2;
    const barCount = parseInt(canvas.width / (barWidth + barSpace));
    const minHeight = 3;
    for (let i = 0; i < barCount; i++) {
      const x = i * (barWidth + barSpace);
      const mapped = Math.pow(i / barCount, 3) * frequency.length;
      const decibels = frequency[parseInt(mapped)];
      const { minDecibels } = analyzer;
      const magnitude = Math.min(Math.max((decibels - minDecibels) / 95, 0), 1);

      const maxHeight = canvas.height - minHeight;
      const height = ((-magnitude * maxHeight) | 0) - minHeight;
      ctx.fillRect(x, canvas.height, barWidth, height);
    }
  };

  render() {
    const { stream, ...otherProps } = this.props;
    const { width, height } = this.state;
    return (
      <div ref={this.container} {...otherProps}>
        <canvas width={width} height={height} ref={this.canvas} />
      </div>
    );
  }
}

Visualizer.propTypes = {
  color: PropTypes.string.isRequired,
};

Visualizer.defaultProps = {
  color: "white",
};

// const Visualizer = ({ stream, ...otherProps }) => {
//   const container = useRef(null);
//   const { clientWidth: width, clientHeight: height } = container.current ?? {};

//   return (
//     <div ref={container} {...otherProps}>
//       <canvas width={width} height={height} />
//     </div>
//   );
// };

export default Visualizer;
