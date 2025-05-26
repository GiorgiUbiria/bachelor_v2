import { createBrowserRouter } from "react-router";
import Root from "./layouts/root";
import Home from "./pages/home";

export const router = createBrowserRouter([
    {
      path: "/",
      Component: Root,
      children: [
        { index: true, Component: Home },
        // { path: "about", Component: About },
        // {
        //   path: "auth",
        //   Component: AuthLayout,
        //   children: [
        //     { path: "login", Component: Login },
        //     { path: "register", Component: Register },
        //   ],
        // },
        // {
        //   path: "products",
        //   children: [
        //     { index: true, Component: ProductsHome },
        //     { path: ":id", Component: ProductDetail },
        //     { path: "search", Component: ProductSearch },
        //   ],
        // },
      ],
    },
  ]);
  