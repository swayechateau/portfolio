class Symbol {
    constructor(x, y, fontSize, canvasHeight) {
        this.characters = 'アァカサタナハマヤャラワガザダバパイィキシチニヒミリヰギジヂビピウゥクスツヌフムユュルグズブヅプエェケセテネヘメレヱゲゼデベペオォコソトノホモヨョロヲゴゾドボポヴッン0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';
        this.x = x;
        this.y = y;
        this.fontSize = fontSize;
        this.text = 'A';
        this.canvasHeight = canvasHeight;
    }

    draw(context) {
        this.text = this.characters.charAt(Math.floor(Math.random() * this.characters.length));
        context.fillText(this.text, this.x * this.fontSize, this.y * this.fontSize);
        if (this.y * this.fontSize > this.canvasHeight && Math.random() > 0.97) {
            this.y = 0;
        } else {
            this.y += 0.9;
        }
    }
}

class Effect {
    constructor(canvasWidth, canvasHeight) {
        this.fontSize = 16;
        this.canvasWidth = canvasWidth;
        this.canvasHeight = canvasHeight;
        this.columns = this.canvasWidth / this.fontSize;
        this.symbols = [];
        this.initialize();
    }

    initialize() {
        for (let i = 0; i < this.columns; i++) {
            this.symbols[i] = new Symbol(i, 0, this.fontSize, this.canvasHeight);
        }
    }

    resize(width, height) {
        this.canvasWidth = width;
        this.canvasHeight = height;
        this.columns = this.canvasWidth / this.fontSize;
        this.symbols = [];
        this.initialize();
    }
}

class Matrix {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = this.canvas.getContext('2d');
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
        this.effect = new Effect(this.canvas.width, this.canvas.height);
        this.last = 0;
        this.fps = 26;
        this.timer = 0;
        this.nextFrame = 1000 / this.fps;
    }

    width(w) {
        this.canvas.width = w;
    }

    height(h) {
        this.canvas.height = h;
    }

    resize() {
        this.effect.resize(this.canvas.width, this.canvas.height);
    }

    animate(time) {
        const deltaTime = time - this.last;
        this.last = time;
        if (this.timer > this.nextFrame) {
            this.ctx.textAlign = 'center';
            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.05)';
            this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
            this.ctx.font = this.effect.fontSize + 'px monospace';
            this.ctx.fillStyle = '#0aff0a';

            this.effect.symbols.forEach(symbol => symbol.draw(this.ctx));
            this.timer = 0;
        } else {
            this.timer += deltaTime;
        }
        requestAnimationFrame(this.animate.bind(this));
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const canvas = document.getElementById('about-matrix');
    const matrix = new Matrix(canvas);
    matrix.animate(0);

    window.addEventListener('resize', () => {
        matrix.width(window.innerWidth);
        matrix.height(window.innerHeight);
        matrix.resize();
    });
});