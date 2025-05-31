import { createBrowserRouter } from "react-router";
import Root from "./layouts/root";
import Home from "./pages/home";
import Login from "./pages/login";
import Register from "./pages/register";
import Products from "./pages/products";
import ProductDetail from "./pages/product-detail";
import Admin from "./pages/admin";

export const router = createBrowserRouter([
  {
    path: "/",
    Component: Root,
    children: [
      { index: true, Component: Home },
      {
        path: "auth",
        children: [
          { path: "login", Component: Login },
          { path: "register", Component: Register },
        ],
      },
      {
        path: "products",
        children: [
          { index: true, Component: Products },
          { path: ":id", Component: ProductDetail },
        ],
      },
      {
        path: "admin",
        Component: Admin,
      },
    ],
  },
]);
