(window.webpackJsonp=window.webpackJsonp||[]).push([[12],{119:function(e,t,n){"use strict";n.d(t,"a",(function(){return s})),n.d(t,"b",(function(){return f}));var r=n(0),a=n.n(r);function o(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function c(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function i(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?c(Object(n),!0).forEach((function(t){o(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):c(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function u(e,t){if(null==e)return{};var n,r,a=function(e,t){if(null==e)return{};var n,r,a={},o=Object.keys(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||(a[n]=e[n]);return a}(e,t);if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(r=0;r<o.length;r++)n=o[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(a[n]=e[n])}return a}var l=a.a.createContext({}),p=function(e){var t=a.a.useContext(l),n=t;return e&&(n="function"==typeof e?e(t):i(i({},t),e)),n},s=function(e){var t=p(e.components);return a.a.createElement(l.Provider,{value:t},e.children)},b={inlineCode:"code",wrapper:function(e){var t=e.children;return a.a.createElement(a.a.Fragment,{},t)}},d=a.a.forwardRef((function(e,t){var n=e.components,r=e.mdxType,o=e.originalType,c=e.parentName,l=u(e,["components","mdxType","originalType","parentName"]),s=p(n),d=r,f=s["".concat(c,".").concat(d)]||s[d]||b[d]||o;return n?a.a.createElement(f,i(i({ref:t},l),{},{components:n})):a.a.createElement(f,i({ref:t},l))}));function f(e,t){var n=arguments,r=t&&t.mdxType;if("string"==typeof e||r){var o=n.length,c=new Array(o);c[0]=d;var i={};for(var u in t)hasOwnProperty.call(t,u)&&(i[u]=t[u]);i.originalType=e,i.mdxType="string"==typeof e?e:r,c[1]=i;for(var l=2;l<o;l++)c[l]=n[l];return a.a.createElement.apply(null,c)}return a.a.createElement.apply(null,n)}d.displayName="MDXCreateElement"},82:function(e,t,n){"use strict";n.r(t),n.d(t,"frontMatter",(function(){return i})),n.d(t,"metadata",(function(){return u})),n.d(t,"toc",(function(){return l})),n.d(t,"default",(function(){return s}));var r=n(3),a=n(7),o=(n(0),n(119)),c=["components"],i={id:"state-db",title:"LevelDB / CouchDB"},u={unversionedId:"operator-guide/state-db",id:"operator-guide/state-db",isDocsHomePage:!1,title:"LevelDB / CouchDB",description:"Once you set the state database of the peer, you cannot change it since the structure is different between using LevelDB and CouchDB you can find more information in the official HLF docs",source:"@site/docs/operator-guide/state-db.md",slug:"/operator-guide/state-db",permalink:"/hlf-operator/docs/operator-guide/state-db",editUrl:"https://github.com/hyperledger-labs/hlf-operator/edit/master/website/docs/operator-guide/state-db.md",version:"current",sidebar:"someSidebar1",previous:{title:"Getting started",permalink:"/hlf-operator/docs/getting-started"},next:{title:"Monitoring",permalink:"/hlf-operator/docs/operator-guide/monitoring"}},l=[{value:"Configuring LevelDB",id:"configuring-leveldb",children:[]},{value:"Configuring CouchDB",id:"configuring-couchdb",children:[]}],p={toc:l};function s(e){var t=e.components,n=Object(a.a)(e,c);return Object(o.b)("wrapper",Object(r.a)({},p,n,{components:t,mdxType:"MDXLayout"}),Object(o.b)("div",{className:"admonition admonition-caution alert alert--warning"},Object(o.b)("div",{parentName:"div",className:"admonition-heading"},Object(o.b)("h5",{parentName:"div"},Object(o.b)("span",{parentName:"h5",className:"admonition-icon"},Object(o.b)("svg",{parentName:"span",xmlns:"http://www.w3.org/2000/svg",width:"16",height:"16",viewBox:"0 0 16 16"},Object(o.b)("path",{parentName:"svg",fillRule:"evenodd",d:"M8.893 1.5c-.183-.31-.52-.5-.887-.5s-.703.19-.886.5L.138 13.499a.98.98 0 0 0 0 1.001c.193.31.53.501.886.501h13.964c.367 0 .704-.19.877-.5a1.03 1.03 0 0 0 .01-1.002L8.893 1.5zm.133 11.497H6.987v-2.003h2.039v2.003zm0-3.004H6.987V5.987h2.039v4.006z"}))),"caution")),Object(o.b)("div",{parentName:"div",className:"admonition-content"},Object(o.b)("p",{parentName:"div"},"Once you set the state database of the peer, you cannot change it since the structure is different between using ",Object(o.b)("strong",{parentName:"p"},"LevelDB")," and ",Object(o.b)("strong",{parentName:"p"},"CouchDB")," you can find more information in ",Object(o.b)("a",{parentName:"p",href:"https://hyperledger-fabric.readthedocs.io/en/release-2.3/couchdb_as_state_database.html"},"the official HLF docs")))),Object(o.b)("h2",{id:"configuring-leveldb"},"Configuring LevelDB"),Object(o.b)("p",null,"In order to configure LevelDB, you need to set the following property in the CRD(Custom resource definition) of the peer:"),Object(o.b)("pre",null,Object(o.b)("code",{parentName:"pre",className:"language-yaml"},"stateDb: leveldb\n")),Object(o.b)("h2",{id:"configuring-couchdb"},"Configuring CouchDB"),Object(o.b)("p",null,"You can configure the world state to be CouchDB by setting the property ",Object(o.b)("inlineCode",{parentName:"p"},"stateDb")," in the CRD of the peer:"),Object(o.b)("pre",null,Object(o.b)("code",{parentName:"pre",className:"language-yaml"},"stateDb: couchdb\n")),Object(o.b)("p",null,"And then you can configure also the username and password:"),Object(o.b)("pre",null,Object(o.b)("code",{parentName:"pre",className:"language-yaml"},"couchdb:\n  externalCouchDB: null\n  password: couchdb\n  user: couchdb\n")),Object(o.b)("p",null,"If you want to configure a custom image for CouchDB, you can set the ",Object(o.b)("inlineCode",{parentName:"p"},"image"),", ",Object(o.b)("inlineCode",{parentName:"p"},"tag"),", and ",Object(o.b)("inlineCode",{parentName:"p"},"pullPolicy")," properties under the ",Object(o.b)("inlineCode",{parentName:"p"},"couchdb")," property:"),Object(o.b)("pre",null,Object(o.b)("code",{parentName:"pre",className:"language-yaml"},"couchdb:\n    image: couchdb\n    pullPolicy: IfNotPresent\n    tag: 3.1.1\n    user: couchdb\n    password: couchdb\n")),Object(o.b)("p",null,"If you wish to use an external CouchDB instance, ",Object(o.b)("a",{parentName:"p",href:"../couchdb/external-couchdb"},"check this page")))}s.isMDXComponent=!0}}]);