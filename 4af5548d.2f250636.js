(window.webpackJsonp=window.webpackJsonp||[]).push([[15],{119:function(e,t,r){"use strict";r.d(t,"a",(function(){return d})),r.d(t,"b",(function(){return y}));var n=r(0),o=r.n(n);function c(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function l(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function a(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?l(Object(r),!0).forEach((function(t){c(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):l(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function i(e,t){if(null==e)return{};var r,n,o=function(e,t){if(null==e)return{};var r,n,o={},c=Object.keys(e);for(n=0;n<c.length;n++)r=c[n],t.indexOf(r)>=0||(o[r]=e[r]);return o}(e,t);if(Object.getOwnPropertySymbols){var c=Object.getOwnPropertySymbols(e);for(n=0;n<c.length;n++)r=c[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(o[r]=e[r])}return o}var u=o.a.createContext({}),p=function(e){var t=o.a.useContext(u),r=t;return e&&(r="function"==typeof e?e(t):a(a({},t),e)),r},d=function(e){var t=p(e.components);return o.a.createElement(u.Provider,{value:t},e.children)},s={inlineCode:"code",wrapper:function(e){var t=e.children;return o.a.createElement(o.a.Fragment,{},t)}},f=o.a.forwardRef((function(e,t){var r=e.components,n=e.mdxType,c=e.originalType,l=e.parentName,u=i(e,["components","mdxType","originalType","parentName"]),d=p(r),f=n,y=d["".concat(l,".").concat(f)]||d[f]||s[f]||c;return r?o.a.createElement(y,a(a({ref:t},u),{},{components:r})):o.a.createElement(y,a({ref:t},u))}));function y(e,t){var r=arguments,n=t&&t.mdxType;if("string"==typeof e||n){var c=r.length,l=new Array(c);l[0]=f;var a={};for(var i in t)hasOwnProperty.call(t,i)&&(a[i]=t[i]);a.originalType=e,a.mdxType="string"==typeof e?e:n,l[1]=a;for(var u=2;u<c;u++)l[u]=r[u];return o.a.createElement.apply(null,l)}return o.a.createElement.apply(null,r)}f.displayName="MDXCreateElement"},85:function(e,t,r){"use strict";r.r(t),r.d(t,"frontMatter",(function(){return a})),r.d(t,"metadata",(function(){return i})),r.d(t,"toc",(function(){return u})),r.d(t,"default",(function(){return d}));var n=r(3),o=r(7),c=(r(0),r(119)),l=["components"],a={id:"develop-chaincode-locally",title:"Develop chaincode locally"},i={unversionedId:"user-guide/develop-chaincode-locally",id:"user-guide/develop-chaincode-locally",isDocsHomePage:!1,title:"Develop chaincode locally",description:"Developing the chaincode locally has the following benefits; for example you can:",source:"@site/docs/user-guide/develop-chaincode-locally.md",slug:"/user-guide/develop-chaincode-locally",permalink:"/hlf-operator/docs/user-guide/develop-chaincode-locally",editUrl:"https://github.com/hyperledger-labs/hlf-operator/edit/master/website/docs/user-guide/develop-chaincode-locally.md",version:"current",sidebar:"someSidebar1",previous:{title:"Register & Enroll users",permalink:"/hlf-operator/docs/user-guide/enroll-users"},next:{title:"Architecture",permalink:"/hlf-operator/docs/chaincode-development/architecture"}},u=[],p={toc:u};function d(e){var t=e.components,r=Object(o.a)(e,l);return Object(c.b)("wrapper",Object(n.a)({},p,r,{components:t,mdxType:"MDXLayout"}),Object(c.b)("p",null,"Developing the chaincode locally has the following benefits; for example you can:"),Object(c.b)("ul",null,Object(c.b)("li",{parentName:"ul"},"Develop the chaincode locally and test it locally."),Object(c.b)("li",{parentName:"ul"},"Debug the chaincode locally."),Object(c.b)("li",{parentName:"ul"},"Improve the installation time by 90% since peers don't need to install and start the chaincode for every change.")))}d.isMDXComponent=!0}}]);