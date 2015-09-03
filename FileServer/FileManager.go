package FileServer

import (
	"duov6.com/FileServer/messaging"
	"duov6.com/common"
	"duov6.com/objectstore/client"
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"github.com/toqueteos/webbrowser"
	"io"
	"io/ioutil"
	"os"
	//"runtime"
	"strings"
)

//var globalConfigs map[string]string

type FileManager struct {
}

type FileData struct {
	Id       string
	FileName string
	Body     string
}

func (f *FileManager) Store(request *messaging.FileRequest) messaging.FileResponse { // store disk on database

	fileResponse := messaging.FileResponse{}

	if len(request.Body) == 0 {

		//WHEN REQUEST COMES FROM A REST INTERFACE

		file, header, err := request.WebRequest.FormFile("file")

		if err != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = err.Error()
		}

		out, err := os.Create(header.Filename)
		if err != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = err.Error()
		}

		// write the content from POST to the file
		_, err = io.Copy(out, file)
		if err != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = err.Error()
		}

		file2, err2 := ioutil.ReadFile(header.Filename)

		if err2 != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = err.Error()
		}

		if checkIfFile(header.Filename) == "xlsx" {
			SaveExcelEntries(header.Filename, request.Parameters["namespace"])
		}

		convertedBody := string(file2[:])
		base64Body := common.EncodeToBase64(convertedBody)

		//store file in the DB as a single file
		obj := FileData{}
		obj.Id = request.Parameters["id"]
		obj.FileName = header.Filename
		obj.Body = base64Body

		headerToken := request.WebRequest.Header.Get("securityToken")

		var extraMap map[string]interface{}
		extraMap = make(map[string]interface{})
		extraMap["File"] = "excelFile"
		fmt.Println(headerToken)
		fmt.Println("Namespace : " + request.Parameters["namespace"])
		fmt.Println("Class : " + request.Parameters["class"])
		//client.GoExtra(headerToken, request.Parameters["namespace"], request.Parameters["class"], extraMap).StoreObject().WithKeyField("Id").AndStoreOne(obj).FileOk()

		fmt.Fprintf(request.WebResponse, "File uploaded successfully : ")
		fmt.Fprintf(request.WebResponse, header.Filename)

		//close the files
		err = out.Close()
		err = file.Close()

		if err != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = err.Error()
		}

		//remove the temporary stored file from the disk
		err2 = os.Remove(header.Filename)

		if err2 != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = err2.Error()
		}

		if err == nil && err2 == nil {
			fileResponse.IsSuccess = true
			fileResponse.Message = "Storing file successfully completed"
		} else {
			fileResponse.IsSuccess = false
			fileResponse.Message = "Storing file was unsuccessfull!" + "\n" + err.Error() + "\n" + err2.Error()
		}

	} else {

		//WHEN REQUEST COMES FROM A NON REST INTERFACE

		convertedBody := string(request.Body[:])
		base64Body := common.EncodeToBase64(convertedBody)

		//store file in the DB as a single file
		obj := FileData{}
		obj.Id = request.Parameters["id"]
		obj.FileName = request.FileName
		obj.Body = base64Body

		client.Go("securityToken", request.Parameters["namespace"], request.Parameters["class"]).StoreObject().WithKeyField("Id").AndStoreOne(obj).FileOk()

		fileResponse.IsSuccess = true
		fileResponse.Message = "Storing file successfully completed"

	}

	return fileResponse
}

func (f *FileManager) Remove(request *messaging.FileRequest) messaging.FileResponse { // remove file from disk and database
	fileResponse := messaging.FileResponse{}
	var saveServerPath string = request.RootSavePath
	file, err := ioutil.ReadFile(saveServerPath + request.FilePath + request.FileName)

	if len(file) > 0 {
		err = os.Remove(saveServerPath + request.FilePath + request.FileName)
	}

	if err == nil {
		fileResponse.IsSuccess = true
		fileResponse.Message = "Deletion of file successfully completed"
	} else {
		fileResponse.IsSuccess = true
		fileResponse.Message = "Deletion of file Aborted"
	}

	obj := FileData{}
	obj.Id = request.Parameters["id"]
	obj.FileName = request.FileName

	client.Go("token", request.Parameters["namespace"], request.Parameters["class"]).StoreObjectWithOperation("delete").WithKeyField("Id").AndStoreOne(obj).Ok()
	fileResponse.IsSuccess = true
	fileResponse.Message = "Deletion of file successfully completed"

	return fileResponse
}

func (f *FileManager) Download(request *messaging.FileRequest) messaging.FileResponse { // save the file to ftp and download via ftp on browser
	fileResponse := messaging.FileResponse{}

	if len(request.Body) == 0 {

	} else {
		var saveServerPath string = request.RootSavePath
		var accessServerPath string = request.RootGetPath

		fmt.Println(request.RootSavePath)
		fmt.Println(request.RootGetPath)

		file := FileData{}
		json.Unmarshal(request.Body, &file)

		temp := common.DecodeFromBase64(file.Body)
		ioutil.WriteFile((saveServerPath + request.FilePath + file.FileName), []byte(temp), 0666)
		err := webbrowser.Open(accessServerPath + request.FilePath + file.FileName)
		if err != nil {
			fileResponse.IsSuccess = false
			fileResponse.Message = "Downloading Failed!" + err.Error()
		} else {
			fileResponse.IsSuccess = true
			fileResponse.Message = "Downloading file successfully completed"
		}

	}

	return fileResponse
}

func SaveExcelEntries(excelFileName string, namespace string) {
	fmt.Println("Inserting Records to Database....")
	rowcount := 0
	colunmcount := 0
	var exceldata []map[string]interface{}
	var colunName []string

	//file read
	xlFile, error := xlsx.OpenFile(excelFileName)

	if error == nil {
		for _, sheet := range xlFile.Sheets {
			rowcount = (sheet.MaxRow - 1)
			colunmcount = sheet.MaxCol
			colunName = make([]string, colunmcount)
			for _, row := range sheet.Rows {
				for j, cel := range row.Cells {
					colunName[j] = cel.String()
				}
				break
			}

			exceldata = make(([]map[string]interface{}), rowcount)

			if error == nil {
				for _, sheet := range xlFile.Sheets {
					for rownumber, row := range sheet.Rows {
						currentRow := make(map[string]interface{})
						if rownumber != 0 {
							exceldata[rownumber-1] = currentRow
							for cellnumber, cell := range row.Cells {
								exceldata[rownumber-1][colunName[cellnumber]] = cell.String()
							}
						}
					}
				}
			}

			Id := colunName[0]

			var extraMap map[string]interface{}
			extraMap = make(map[string]interface{})
			extraMap["File"] = "exceldata"
			fmt.Println("Namespace : " + namespace)
			fmt.Println("filename : " + getExcelFileName(excelFileName) /*+ strings.ToLower(sheet.Name)*/)
			client.GoExtra("token", namespace, getExcelFileName(excelFileName) /*+strings.ToLower(sheet.Name)*/, extraMap).StoreObject().WithKeyField(Id).AndStoreMapInterface(exceldata).Ok()
			//client.GoExtra("token", namespace, getExcelFileName(excelFileName), extraMap).StoreObject().WithKeyField(Id).AndStoreMapInterface(exceldata).Ok()

		}

	}
	return
}

func checkIfFile(params string) (fileType string) {

	var tempArray []string
	tempArray = strings.Split(params, ".")
	if len(tempArray) > 1 {
		fileType = tempArray[len(tempArray)-1]
	} else {
		fileType = "NAF"
	}
	return
}

func getExcelFileName(path string) (fileName string) {
	subsets := strings.Split(path, "\\")
	subfilenames := strings.Split(subsets[len(subsets)-1], ".")
	fileName = subfilenames[0]
	return
}
