import * as React from "react";
import md5 = require("md5");
import { AxiosError, AxiosResponse } from "axios";

export const Gravatar = (props: {email?: string}) => {
  const hash = props.email ? md5(props.email) : "";
  return <img className="ui avatar image" src={ "https://www.gravatar.com/avatar/" + hash } />;
};

export const MultiLineText = (props: {text?: string}) => {
  if (!props.text) {
    return <p></p>;
  }

  return <div>{props.text.split("\n").map((item, i) =>
   <span>{item}<br/></span>
  )}</div>;
};

export const DisplayError = (props: {error?: Error}) => {
  if (!props.error) {
    return <div></div>;
  }

  return <div className="ui negative message">
            <div className="header">
              Oops, an error ocurred...
            </div>
            <p>{ props.error.message }</p>
         </div>;
};