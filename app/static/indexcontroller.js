var app = new Vue({
    delimiters: ['[[ ', ' ]]'],
    el: '#app',
    data: {
        activeIndex: '/'
    },
    methods: {

        handleSelect: function (key, keyPath) {
            if (key === '#' || key === '/') {
                this.activeIndex = '/';
                window.location.href = '/'
            }
            if (key === 'search') {
                this.activeIndex = "search"
                window.location.href = "search"
            }
        }
    },
    created: function () {
    }
})
export default app